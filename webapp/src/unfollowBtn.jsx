import React, { useContext } from 'react';
import axios from 'axios';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import MattermostContext from './contexts/MattermostContext';
import { extractCsrfToken } from './utilities/cookies';

import StoreContext from './contexts/StoreContext';

function UnfollowBtn() {
    const queryClient = useQueryClient();
    const mmProps = useContext(MattermostContext);
    const store = useContext(StoreContext);
    const state = store.getState();
    let { SiteURL } = state.entities.general.config;
    if (SiteURL === '') {
        SiteURL = 'http://localhost:8065';
    }

    const mutation = useMutation({
        mutationFn: () => axios.delete(`${SiteURL}/plugins/com.tcg.followers/follow?follow_id=${mmProps.user.id}`, { headers: { 'X-CSRF-Token': extractCsrfToken() } }),
        onSuccess: () => {
            queryClient.setQueryData(['followedUsers'], (oldQueryData) => {
                return oldQueryData.filter((id) => id !== mmProps.user.id);
            });
        },
        onError: () => {
            // do something?
        },
    });

    return (
        <div
            className='popover__row'
            style={{ paddingTop: '0 !important' }}
        >
            <button
                type='button'
                className='btn'
                style={{ width: '100%' }}
                onClick={mutation.mutate}
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
