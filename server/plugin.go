package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/shared/i18n"
	"github.com/mattermost/mattermost/server/v8/channels/utils"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

func (p *Plugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	// check to see if anyone is following the user who posted the message
	// if a user is following the user who posted the message, send a push notification
	followers, err := p.ListFollowedBy(post.UserId)
	if err != nil {
		p.API.LogError("Error getting followers", "error", err.Error())
		return
	}

	channel, appErr := p.API.GetChannel(post.ChannelId)
	if appErr != nil {
		p.API.LogError("Error getting channel", "error", appErr.Error())
		return
	}

	user, appErr := p.API.GetUser(post.UserId)
	if appErr != nil {
		p.API.LogError("Error getting user", "error", appErr.Error())
		return
	}

	cfg := p.API.GetConfig()

	msg := &model.PushNotification{
		Category:     model.CategoryCanReply,
		Version:      model.PushMessageV2,
		Type:         model.PushTypeMessage,
		TeamId:       channel.TeamId,
		ChannelId:    channel.Id,
		PostId:       post.Id,
		RootId:       post.RootId,
		SenderId:     post.UserId,
		SenderName:   user.Username,
		IsCRTEnabled: true,  // is threads enabled (?)
		IsIdLoaded:   false, // indicates that the user has received a new PM (?)
	}

	userLocale := i18n.GetUserTranslations(user.Locale)
	if post.RootId != "" {
		props := map[string]any{"channelName": channel.DisplayName}
		msg.ChannelName = userLocale("api.push_notification.title.collapsed_threads", props)

		if channel.Type == model.ChannelTypeDirect {
			msg.ChannelName = userLocale("api.push_notification.title.collapsed_threads_dm")
		}
	}

	if ou, ok := post.GetProp("override_username").(string); ok && *cfg.ServiceSettings.EnablePostUsernameOverride {
		msg.OverrideUsername = ou
		msg.SenderName = ou
	}

	if oi, ok := post.GetProp("override_icon_url").(string); ok && *cfg.ServiceSettings.EnablePostIconOverride {
		msg.OverrideIconURL = oi
	}

	if fw, ok := post.GetProp("from_webhook").(string); ok {
		msg.FromWebhook = fw
	}

	postMessage := post.Message
	stripped, err := utils.StripMarkdown(postMessage)
	if err != nil {
		// c.Logger().Warn("Failed parse to markdown", mlog.String("post_id", post.Id), mlog.Err(err))
		p.API.LogError("Failed parse to markdown", "post_id", post.Id, "error", err.Error())
	} else {
		postMessage = stripped
	}
	for _, attachment := range post.Attachments() {
		if attachment.Fallback != "" {
			postMessage += "\n" + attachment.Fallback
		}
	}

	msg.Message = postMessage

	for _, follower := range followers {
		p.SendPushNotification(follower, msg)
	}
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	// p.API.LogInfo("************************ Requested URL Path: " + r.URL.Path)
	switch r.URL.Path {
	case "/hello":
		_, wErr := w.Write([]byte("Hello, world!"))
		if wErr != nil {
			p.API.LogError("Failed to write response", "error", wErr.Error())
		}
	case "/all_follows":
		switch r.Method {
		case http.MethodGet:
			data, appErr := p.AllFollows()
			if appErr != nil {
				http.Error(w, appErr.Error(), http.StatusInternalServerError)
				return
			}

			// Convert data to []byte
			dataBytes, err := json.Marshal(data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Respond with the data
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, wErr := w.Write(dataBytes)
			if wErr != nil {
				p.API.LogError("Failed to write response", "error", wErr.Error())
			}
		default:
			http.NotFound(w, r)
			return
		}
	case "/delete_all_follows":
		switch r.Method {
		case http.MethodGet:
			if appErr := p.KVDeleteAll(); appErr != nil {
				http.Error(w, appErr.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
			return
		}
	case "/follow":
		switch r.Method {
		case http.MethodPost:
			// Handle POST request as previously defined
			var postData struct {
				FollowID string `json:"follow_id"` // Assuming your postData has a field for follow_id
			}

			mattermostUserID := r.Header.Get("Mattermost-User-Id") // Extract the Mattermost-User-Id header
			if mattermostUserID == "" {
				http.Error(w, "Mattermost-User-Id header missing", http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}

			err = json.Unmarshal(body, &postData)
			if err != nil {
				http.Error(w, "Error parsing request body", http.StatusBadRequest)
				return
			}

			// Assume you extract or define the key for KVSet here
			if appErr := p.Follow(mattermostUserID, postData.FollowID); appErr != nil {
				http.Error(w, appErr.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		case http.MethodDelete:
			// Handle DELETE request
			followID := r.URL.Query().Get("follow_id")
			if followID == "" {
				http.Error(w, "follow_id is required", http.StatusBadRequest)
				return
			}

			mattermostUserID := r.Header.Get("Mattermost-User-Id") // Extract the Mattermost-User-Id header
			if mattermostUserID == "" {
				http.Error(w, "Mattermost-User-Id header missing", http.StatusBadRequest)
				return
			}

			if appErr := p.Unfollow(mattermostUserID, followID); appErr != nil {
				http.Error(w, appErr.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Key deleted successfully")
			return
		case http.MethodGet:
			mattermostUserID := r.Header.Get("Mattermost-User-Id") // Extract the Mattermost-User-Id header
			if mattermostUserID == "" {
				http.Error(w, "Mattermost-User-Id header missing", http.StatusBadRequest)
				return
			}

			data, appErr := p.ListFollows(mattermostUserID)
			if appErr != nil {
				http.Error(w, appErr.Error(), http.StatusForbidden)
				return
			}

			// Check if data exists for the key
			if data == nil {
				http.NotFound(w, r)
				return
			}

			// Convert data to JSON
			jsonData, err := json.Marshal(data)
			if err != nil {
				http.Error(w, "Failed to encode data as JSON", http.StatusInternalServerError)
				return
			}

			// Respond with the JSON data
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, wErr := w.Write(jsonData)
			if wErr != nil {
				p.API.LogError("Failed to write response", "error", wErr.Error())
			}
		default:
			http.NotFound(w, r)
			return
		}
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) Follow(currentUserID, followerID string) error {
	currentFollows, err := p.ListFollows(currentUserID)
	if err != nil {
		return err
	}

	currentFollows = append(currentFollows, followerID)
	err = p.KVSet(fmt.Sprintf("%s:following", currentUserID), currentFollows)
	if err != nil {
		return err
	}

	followerFollowedBy, err := p.ListFollowedBy(followerID)
	if err != nil {
		return err
	}

	followerFollowedBy = append(followerFollowedBy, currentUserID)
	err = p.KVSet(fmt.Sprintf("%s:followed_by", followerID), followerFollowedBy)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) Unfollow(currentUserID, followerID string) error {
	currentFollows, err := p.ListFollows(currentUserID)
	if err != nil {
		return err
	}

	for i, id := range currentFollows {
		if id == followerID {
			currentFollows = append(currentFollows[:i], currentFollows[i+1:]...)
			break
		}
	}
	err = p.KVSet(fmt.Sprintf("%s:following", currentUserID), currentFollows)
	if err != nil {
		return err
	}

	followerFollowedBy, err := p.ListFollowedBy(followerID)
	if err != nil {
		return err
	}

	for i, id := range followerFollowedBy {
		if id == currentUserID {
			followerFollowedBy = append(followerFollowedBy[:i], followerFollowedBy[i+1:]...)
			break
		}
	}
	err = p.KVSet(fmt.Sprintf("%s:followed_by", followerID), followerFollowedBy)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) ListFollows(currentUserID string) ([]string, error) {
	data, appErr := p.KVGet(fmt.Sprintf("%s:following", currentUserID))
	if appErr != nil {
		return nil, appErr
	}

	if len(data) == 0 {
		return []string{}, nil
	}

	var follows []string
	if jsonErr := json.Unmarshal(data, &follows); jsonErr != nil {
		return nil, jsonErr
	}
	return follows, nil
}

func (p *Plugin) AllFollows() (map[string][]string, error) {
	keys, err := p.API.KVList(0, 100)
	if err != nil {
		return nil, err
	}

	// iterate over the keys and store the keys and the values in a map
	followData := make(map[string][]string)
	for _, key := range keys {
		if strings.HasSuffix(key, ":following") {
			data, err := p.API.KVGet(key)
			if err != nil {
				return nil, err
			}

			// Use a slice to hold the decoded data
			var followers []string
			jsonErr := json.Unmarshal(data, &followers)
			if jsonErr != nil {
				return nil, jsonErr
			}

			followData[key] = followers
		}
	}

	return followData, nil
}

func (p *Plugin) ListFollowedBy(userID string) ([]string, error) {
	data, err := p.KVGet(fmt.Sprintf("%s:followed_by", userID))
	if err != nil {
		return nil, err // Return immediately if there's an error fetching the data
	}
	if data == nil {
		return []string{}, nil // If data is nil, return an empty slice without error
	}

	var follows []string
	jsonErr := json.Unmarshal(data, &follows)
	if jsonErr != nil {
		return nil, jsonErr // Return an error if unmarshaling fails
	}
	return follows, nil // Return the successfully unmarshaled slice
}

func (p *Plugin) KVSet(key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	mErr := p.API.KVSet(key, data)
	if mErr != nil {
		return mErr.Unwrap()
	}

	return nil
}

func (p *Plugin) KVDelete(key string) error {
	mErr := p.API.KVDelete(key)
	if mErr != nil {
		return mErr.Unwrap()
	}

	return nil
}

func (p *Plugin) KVDeleteAll() error {
	mErr := p.API.KVDeleteAll()
	if mErr != nil {
		return mErr.Unwrap()
	}

	return nil
}

func (p *Plugin) KVGet(key string) ([]byte, error) {
	b, err := p.API.KVGet(key)
	if err != nil {
		return b, err.Unwrap()
	}

	return b, nil
}

func (p *Plugin) SendPushNotification(userID string, notification *model.PushNotification) {
	err := p.API.SendPushNotification(notification, userID)
	if err != nil {
		p.API.LogInfo("********************** Error sending push notification " + err.Error())
	}
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
