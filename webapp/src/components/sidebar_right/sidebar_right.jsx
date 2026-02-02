import React, {useState, useEffect} from 'react';
import {FormattedMessage} from 'react-intl';

import BadgeItem from '../badge_item/badge_item';

const SidebarRight = ({siteURL, currentUserId}) => {
    const [badges, setBadges] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(false);

    useEffect(() => {
        if (!currentUserId) {
            return;
        }

        const fetchBadges = async () => {
            try {
                setLoading(true);
                setError(false);
                const resp = await fetch(
                    `${siteURL}/plugins/com.mattermost.badges/api/v1/getUserBadges/${currentUserId}`,
                    {credentials: 'include'},
                );
                if (!resp.ok) {
                    throw new Error('Failed to fetch badges');
                }
                const data = await resp.json();
                setBadges(data || []);
            } catch (e) {
                setError(true);
            } finally {
                setLoading(false);
            }
        };

        fetchBadges();
    }, [currentUserId, siteURL]);

    if (loading) {
        return (
            <div style={{padding: '16px', textAlign: 'center', color: 'rgba(var(--center-channel-color-rgb), 0.64)'}}>
                <FormattedMessage
                    id='SidebarRight.loading'
                    defaultMessage='Loading...'
                />
            </div>
        );
    }

    if (error) {
        return (
            <div style={{padding: '16px', textAlign: 'center', color: 'var(--error-text)'}}>
                <FormattedMessage
                    id='SidebarRight.error'
                    defaultMessage='Failed to load badges.'
                />
            </div>
        );
    }

    if (badges.length === 0) {
        return (
            <div style={{padding: '16px', textAlign: 'center', color: 'rgba(var(--center-channel-color-rgb), 0.64)'}}>
                <FormattedMessage
                    id='SidebarRight.noBadges'
                    defaultMessage="You don't have any badges yet."
                />
            </div>
        );
    }

    return (
        <div>
            {badges.map((badge, index) => (
                <BadgeItem
                    key={badge.id || index}
                    badge={badge}
                />
            ))}
        </div>
    );
};

export default SidebarRight;
