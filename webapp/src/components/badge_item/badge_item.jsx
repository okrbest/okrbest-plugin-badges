import React from 'react';
import {FormattedMessage} from 'react-intl';

const BadgeItem = ({badge}) => {
    const imageContent = badge.image ? (
        <span style={{fontSize: '24px', marginRight: '12px'}}>{badge.image}</span>
    ) : (
        <span style={{
            width: '32px',
            height: '32px',
            borderRadius: '50%',
            backgroundColor: '#ddd',
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginRight: '12px',
            fontSize: '16px',
        }}>
            {'🏅'}
        </span>
    );

    return (
        <div style={{
            display: 'flex',
            alignItems: 'flex-start',
            padding: '12px 16px',
            borderBottom: '1px solid rgba(var(--center-channel-color-rgb), 0.08)',
        }}>
            {imageContent}
            <div style={{flex: 1, minWidth: 0}}>
                <div style={{
                    fontWeight: 600,
                    fontSize: '14px',
                    color: 'var(--center-channel-color)',
                }}>
                    {badge.name}
                </div>
                {badge.description && (
                    <div style={{
                        fontSize: '12px',
                        color: 'rgba(var(--center-channel-color-rgb), 0.64)',
                        marginTop: '2px',
                    }}>
                        {badge.description}
                    </div>
                )}
                {badge.granted_by_username && (
                    <div style={{
                        fontSize: '11px',
                        color: 'rgba(var(--center-channel-color-rgb), 0.48)',
                        marginTop: '4px',
                    }}>
                        <FormattedMessage
                            id='BadgeItem.grantedBy'
                            defaultMessage='Granted by {username}'
                            values={{username: badge.granted_by_username}}
                        />
                    </div>
                )}
                {badge.reason && (
                    <div style={{
                        fontSize: '11px',
                        color: 'rgba(var(--center-channel-color-rgb), 0.48)',
                        marginTop: '2px',
                    }}>
                        <FormattedMessage
                            id='BadgeItem.reason'
                            defaultMessage='Reason: {reason}'
                            values={{reason: badge.reason}}
                        />
                    </div>
                )}
            </div>
        </div>
    );
};

export default BadgeItem;
