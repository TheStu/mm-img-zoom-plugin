users following users
- some kind of dropdown or button to initiate the follow
- store that ... in the local DB? I think so
- button to unfollow

API endpoints
- create follow
- delete follow
- get all follows
- for each 'follow', we need to track it for the follower and the followee

sending notifications
- when anyone posts, consult the table to see if anyone follows them
- if followers found, send a msg

So pieces needed:
- user interface mods
- hook on user posts -> send push notifications
