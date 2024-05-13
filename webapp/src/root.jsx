import { useQuery } from '@tanstack/react-query';
import { useContext } from 'react';

import FollowBtn from './followBtn';
import { fetchData } from './fetch/followers';
import MattermostContext from './contexts/MattermostContext';

import UnfollowBtn from './unfollowBtn';
import StoreContext from './contexts/StoreContext';

function Root() {
    const mmProps = useContext(MattermostContext);
    const store = useContext(StoreContext);
    const state = store.getState();
    let { SiteURL } = state.entities.general.config;
    if (SiteURL === '') {
        SiteURL = 'http://localhost:8065';
    }

    const { data, error, isLoading } = useQuery(['followedUsers'], () => fetchData(SiteURL));

    if (isLoading) {
        return 'Loading...';
    }

    if (error) {
        return <div>{`Error: ${error.message}`}</div>;
    }

    if (data.includes(mmProps.user.id)) {
        return <UnfollowBtn/>;
    }

    return <FollowBtn/>;
}

export default Root;
