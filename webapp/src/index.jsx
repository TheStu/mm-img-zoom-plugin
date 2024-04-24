// import {Store, Action} from 'redux';

// import {GlobalState} from '@mattermost/types/lib/store';

// import manifest from '@/manifest';

// import {PluginRegistry} from '@/types/mattermost-webapp';

// import FollowBtn from './followBtn';

// export default class Plugin {
//     // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
//     public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
//         // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
//         registry.registerPopoverUserActionsComponent(FollowBtn);
//     }
// }

// declare global {
//     interface Window {
//         registerPlugin(pluginId: string, plugin: Plugin): void
//     }
// }

// window.registerPlugin(manifest.id, new Plugin());
import {
    QueryClient,
    QueryClientProvider,
} from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

import manifest from '@/manifest';

import Root from './root';
import StoreContext from './contexts/StoreContext';
import MattermostContext from './contexts/MattermostContext';

class Plugin {
    async initialize(registry, store) {
        // See https://developers.mattermost.com/extend/plugins/webapp/reference/
        registry.registerPopoverUserActionsComponent((props) => (
            <StoreContext.Provider value={store}>
                <MattermostContext.Provider value={props}>
                    <QueryClientProvider client={new QueryClient()}>
                        <Root/>
                        <ReactQueryDevtools initialIsOpen={false}/>
                    </QueryClientProvider>
                </MattermostContext.Provider>
            </StoreContext.Provider>
        ));
    }
}

// window.registerPlugin = (pluginId, plugin) => {
//     window.registerPlugin(pluginId, plugin);
// };

window.registerPlugin(manifest.id, new Plugin());
