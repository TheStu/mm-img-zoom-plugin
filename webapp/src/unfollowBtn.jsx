import React, { useContext } from 'react';
import axios from 'axios';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import MattermostContext from './contexts/MattermostContext';
import { extractCsrfToken } from './utilities/cookies';

function UnfollowBtn() {
    const queryClient = useQueryClient();
    const mmProps = useContext(MattermostContext);

    const mutation = useMutation({
        mutationFn: (newFollow) => axios.delete(`http://localhost:8065/plugins/com.tcg.followers/follow?follow_id=${mmProps.user.id}`, newFollow, { headers: { 'X-CSRF-Token': extractCsrfToken() } }),
        onSuccess: () => {
            queryClient.setQueryData(['followedUsers'], (oldQueryData) => {
                return oldQueryData.filter((id) => id !== mmProps.user.id);
            });
        },
        onError: () => {
            // do something?
        },
    });

    const sendUnfollowRequest = () => {
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
                onClick={sendUnfollowRequest}
                disabled={mutation.isLoading}
            >
                {mutation.isLoading ? (
                    <span>{'Loading...'}</span>
                ) : (
                    <>
                        <i
                            className='icon icon-close'
                            style={{ marginRight: '0 !important' }}
                        />
                        <span>{'Unfollow'}</span>
                    </>
                )}
            </button>
        </div>
    );
}

export default UnfollowBtn;
