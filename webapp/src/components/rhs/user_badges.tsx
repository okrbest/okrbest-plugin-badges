import React from 'react';

import {UserProfile} from 'mattermost-redux/types/users';
import {systemEmojis} from 'mattermost-redux/actions/emojis';

import {injectIntl, IntlShape} from 'react-intl';

import {BadgeID, UserBadge} from '../../types/badges';
import Client from '../../client/api';

import {RHSState} from 'types/general';
import {IMAGE_TYPE_EMOJI, RHS_STATE_DETAIL} from '../../constants';

import UserBadgeRow from './user_badge_row';
import RHSScrollbars from './rhs_scrollbars';

import './user_badges.scss';

type Props = {
    intl: IntlShape;
    isCurrentUser: boolean;
    user: UserProfile | null;
    actions: {
        setRHSView: (view: RHSState) => void;
        setRHSBadge: (badge: BadgeID | null) => void;
        getCustomEmojisByName: (names: string[]) => void;
    };
}

type State = {
    loading: boolean;
    badges?: UserBadge[];
}

class UserBadges extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props);

        this.state = {
            loading: true,
        };
    }

    componentDidMount() {
        if (!this.props.user) {
            return;
        }
        const c = new Client();
        c.getUserBadges(this.props.user.id).then((badges) => {
            this.setState({badges, loading: false});
        });
    }

    componentDidUpdate(prevProps: Props, prevState: State) {
        if (this.state.badges !== prevState.badges) {
            const names: string[] = [];
            this.state.badges?.forEach((badge) => {
                if (badge.image_type === IMAGE_TYPE_EMOJI) {
                    names.push(badge.image);
                }
            });
            const toLoad = names.filter((v) => !systemEmojis.has(v));
            this.props.actions.getCustomEmojisByName(toLoad);
        }
        if (this.props.user?.id === prevProps.user?.id) {
            return;
        }

        if (!this.props.user) {
            return;
        }

        const c = new Client();
        if (!this.state.loading) {
            this.setState({loading: true});
        }

        c.getUserBadges(this.props.user.id).then((badges) => {
            this.setState({badges, loading: false});
        });
    }

    onBadgeClick = (badge: UserBadge) => {
        this.props.actions.setRHSBadge(badge.id);
        this.props.actions.setRHSView(RHS_STATE_DETAIL);
    }

    render() {
        const {intl} = this.props;

        if (!this.props.user) {
            return (<div>{intl.formatMessage({id: 'UserBadges.notFound', defaultMessage: 'User not found.'})}</div>);
        }

        if (this.state.loading) {
            return (<div>{intl.formatMessage({id: 'Common.loading', defaultMessage: 'Loading...'})}</div>);
        }

        if (!this.state.badges || this.state.badges.length === 0) {
            return (<div>{intl.formatMessage({id: 'UserBadges.noBadges', defaultMessage: 'No badges yet.'})}</div>);
        }

        const content = this.state.badges.map((badge) => {
            return (
                <UserBadgeRow
                    isCurrentUser={this.props.isCurrentUser}
                    key={badge.time}
                    badge={badge}
                    onClick={this.onBadgeClick}
                />
            );
        });

        let title = intl.formatMessage({id: 'UserBadges.myBadges', defaultMessage: 'My badges'});
        if (!this.props.isCurrentUser) {
            title = intl.formatMessage({id: 'UserBadges.userBadges', defaultMessage: "@{username}'s badges"}, {username: this.props.user.username});
        }
        return (
            <div className='UserBadges'>
                <div><b>{title}</b></div>
                <RHSScrollbars>{content}</RHSScrollbars>
            </div>
        );
    }
}

export default injectIntl(UserBadges);
