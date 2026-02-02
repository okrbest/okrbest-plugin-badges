import {Store} from 'redux';

import {GlobalState} from 'mattermost-redux/types/store';

import {GenericAction} from 'mattermost-redux/types/actions';

import {getConfig} from 'mattermost-redux/selectors/entities/general';

import {getCurrentUser} from 'mattermost-redux/selectors/entities/users';

import React from 'react';

import {createIntl, createIntlCache} from 'react-intl';

import {openAddSubscription, openCreateBadge, openCreateType, openRemoveSubscription, setRHSView, setShowRHSAction} from 'actions/actions';

import RHS from 'components/rhs';

import ChannelHeaderButton from 'components/channel_header_button';

import Reducer from './reducers';

import manifest from './manifest';

// eslint-disable-next-line import/no-unresolved
import {PluginRegistry} from './types/mattermost-webapp';
import BadgeList from './components/user_popover/';
import {RHS_STATE_ALL} from './constants';
import {getMessages} from './i18n';

export default class Plugin {
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, GenericAction>) {
        const state = store.getState();
        const currentUser = getCurrentUser(state);
        const locale = (currentUser?.locale || 'ko').split('-')[0];

        const cache = createIntlCache();
        const intl = createIntl({
            locale,
            messages: getMessages(locale),
        }, cache);

        registry.registerReducer(Reducer);

        registry.registerPopoverUserAttributesComponent(BadgeList);

        const {showRHSPlugin, toggleRHSPlugin} = registry.registerRightHandSidebarComponent(
            RHS,
            intl.formatMessage({id: 'SidebarRight.title', defaultMessage: 'My Badges'}),
        );
        store.dispatch(setShowRHSAction(() => store.dispatch(showRHSPlugin)));

        const toggleRHS = () => {
            store.dispatch(setRHSView(RHS_STATE_ALL));
            store.dispatch(toggleRHSPlugin);
        };

        registry.registerChannelHeaderButtonAction(
            <ChannelHeaderButton/>,
            toggleRHS,
            intl.formatMessage({id: 'Plugin.channelHeader.badges', defaultMessage: 'Badges'}),
            intl.formatMessage({id: 'Plugin.channelHeader.tooltip', defaultMessage: 'Open your badges'}),
        );

        if (registry.registerAppBarComponent) {
            const siteUrl = getConfig(store.getState())?.SiteURL || '';
            const iconURL = `${siteUrl}/plugins/${manifest.id}/public/app-bar-icon.png`;
            registry.registerAppBarComponent(
                iconURL,
                toggleRHS,
                intl.formatMessage({id: 'Plugin.appBar.tooltip', defaultMessage: 'Open your badges'}),
            );
        }

        registry.registerMainMenuAction(
            intl.formatMessage({id: 'Menu.createBadge', defaultMessage: 'Create badge'}),
            () => {
                store.dispatch(openCreateBadge() as any);
            },
            null,
        );
        registry.registerMainMenuAction(
            intl.formatMessage({id: 'Menu.createBadgeType', defaultMessage: 'Create badge type'}),
            () => {
                store.dispatch(openCreateType() as any);
            },
            null,
        );

        registry.registerChannelHeaderMenuAction(
            intl.formatMessage({id: 'Menu.addSubscription', defaultMessage: 'Add badge subscription'}),
            () => {
                store.dispatch(openAddSubscription() as any);
            },
        );
        registry.registerChannelHeaderMenuAction(
            intl.formatMessage({id: 'Menu.removeSubscription', defaultMessage: 'Remove badge subscription'}),
            () => {
                store.dispatch(openRemoveSubscription() as any);
            },
        );
    }
}

declare global {
    interface Window {
        registerPlugin(id: string, plugin: Plugin): void;
    }
}

window.registerPlugin(manifest.id, new Plugin());
