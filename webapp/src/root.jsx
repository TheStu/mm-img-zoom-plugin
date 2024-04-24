import { useQuery } from '@tanstack/react-query';
import { useContext } from 'react';

import FollowBtn from './followBtn';
import { fetchData } from './fetch/followers';
import MattermostContext from './contexts/MattermostContext';

import UnfollowBtn from './unfollowBtn';

function Root() {
    const { data, error, isLoading } = useQuery(['followedUsers'], fetchData);
    const mmProps = useContext(MattermostContext);

    if (isLoading) {
        return '';
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
