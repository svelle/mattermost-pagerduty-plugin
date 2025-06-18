// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import type {Store, Action} from 'redux';

import type {GlobalState} from '@mattermost/types/store';

import PagerDutySidebar from './components/sidebar/sidebar';

import manifest from '@/manifest';
import type {PluginRegistry} from '@/types/mattermost-webapp';

// Define the icon as an inline SVG component
const Icon = () => (
    <svg
        width='18'
        height='18'
        viewBox='0 0 64 64'
        xmlns='http://www.w3.org/2000/svg'
        style={{fill: 'currentColor'}}
    >
        <rect
            width='64'
            height='64'
            rx='8'
            fill='#06AC38'
        />
        <path
            d='M 16 16 L 32 16 Q 40 16 44 20 Q 48 24 48 32 Q 48 40 44 44 Q 40 48 32 48 L 24 48 L 24 56 L 16 56 Z M 24 24 L 24 40 L 32 40 Q 36 40 38 38 Q 40 36 40 32 Q 40 28 38 26 Q 36 24 32 24 Z'
            fill='white'
        />
    </svg>
);

export default class Plugin {
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // Register the RHS component
        const {toggleRHSPlugin} = registry.registerRightHandSidebarComponent(
            PagerDutySidebar,
            'PagerDuty',
        );

        // Register channel header button
        registry.registerChannelHeaderButtonAction(
            <Icon/>,
            () => store.dispatch(toggleRHSPlugin),
            'View PagerDuty on-call schedules',
        );
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void;
    }
}

window.registerPlugin(manifest.id, new Plugin());
