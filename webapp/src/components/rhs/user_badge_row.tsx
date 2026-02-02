import React from 'react';

import Client4 from 'mattermost-redux/client/client4';

import {useIntl} from 'react-intl';

import {UserBadge} from '../../types/badges';
import BadgeImage from '../utils/badge_image';
import {markdown} from 'utils/markdown';

import './user_badge_row.scss';

type Props = {
    badge: UserBadge;
    isCurrentUser: boolean;
    onClick: (badge: UserBadge) => void;
}

const UserBadgeRow: React.FC<Props> = ({badge, onClick, isCurrentUser}: Props) => {
    const intl = useIntl();
    const time = new Date(badge.time);
    let reason = null;
    if (badge.reason) {
        reason = (<div className='badge-user-reason'>{intl.formatMessage({id: 'UserBadgeRow.reason', defaultMessage: 'Why? {reason}'}, {reason: badge.reason})}</div>);
    }
    let setStatus = null;
    if (isCurrentUser && badge.image_type === 'emoji') {
        setStatus = (
            <div className='user-badge-set-status'>
                <a
                    onClick={() => {
                        const c = new Client4();
                        c.updateCustomStatus({emoji: badge.image, text: badge.name});
                    }}
                >
                    {intl.formatMessage({id: 'UserBadgeRow.setStatus', defaultMessage: 'Set status to this badge'})}
                </a>
            </div>
        );
    }
    return (
        <div className='UserBadgesRow'>
            <a onClick={() => onClick(badge)}>
                <span className='user-badge-icon'>
                    <BadgeImage
                        badge={badge}
                        size={32}
                    />
                </span>
            </a>
            <div className='user-badge-text'>
                <div className='user-badge-name'>{badge.name}</div>
                <div className='user-badge-description'>{markdown(badge.description)}</div>
                {reason}
                <div className='user-badge-type'>{intl.formatMessage({id: 'UserBadgeRow.type', defaultMessage: 'Type: {typeName}'}, {typeName: badge.type_name})}</div>
                <div className='user-badge-granted-by'>{intl.formatMessage({id: 'UserBadgeRow.grantedBy', defaultMessage: 'Granted by: {username}'}, {username: badge.granted_by_name})}</div>
                <div className='user-badge-granted-at'>{intl.formatMessage({id: 'UserBadgeRow.grantedAt', defaultMessage: 'Granted at: {date}'}, {date: intl.formatDate(time, {year: 'numeric', month: 'long', day: 'numeric'})})}</div>
                {setStatus}
            </div>
        </div>
    );
};

export default UserBadgeRow;
