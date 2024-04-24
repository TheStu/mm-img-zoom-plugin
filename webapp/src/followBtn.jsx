import React, { useContext } from 'react';
import axios from 'axios';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import MattermostContext from './contexts/MattermostContext';
import { extractCsrfToken } from './utilities/cookies';

function FollowBtn() {
    const queryClient = useQueryClient();
    const mmProps = useContext(MattermostContext);

    const mutation = useMutation({
        mutationFn: (newFollow) => axios.post('http://localhost:8065/plugins/com.tcg.followers/follow', newFollow, { headers: { 'X-CSRF-Token': extractCsrfToken() } }),
        onSuccess: () => {
            queryClient.setQueryData(['followedUsers'], (oldQueryData) => {
                return [...oldQueryData, mmProps.user.id];
            });
        },
        onError: () => {
            // do something?
        },
    });

    const sendFollowRequest = () => {
        const userData = {
            follow_id: mmProps.user.id,
        };
        mutation.mutate(userData);
    };

    return (
        <div
            className='popover__row'
            style={{ paddingTop: '0 !important' }}
        >
            <button
                type='button'
                className='btn'
                style={{ width: '100%' }}
                onClick={sendFollowRequest}
                disabled={mutation.isLoading}
            >
                {mutation.isLoading ? (
                    <span>{'Loading...'}</span>
                ) : (
                    <>
                        <i
                            className='icon icon-plus'
                            style={{ marginRight: '0 !important' }}
                        />
                        <span>{'Follow'}</span>
                    </>
                )}
            </button>
        </div>
    );
}

export default FollowBtn;
