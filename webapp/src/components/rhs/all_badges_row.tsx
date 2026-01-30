import React from 'react';

import {useIntl} from 'react-intl';

import {AllBadgesBadge} from '../../types/badges';
import BadgeImage from '../utils/badge_image';
import {markdown} from 'utils/markdown';

import './all_badges_row.scss';

type Props = {
    badge: AllBadgesBadge;
    onClick: (badge: AllBadgesBadge) => void;
}

const AllBadgesRow: React.FC<Props> = ({badge, onClick}: Props) => {
    const intl = useIntl();

    let grantedText: string;
    if (badge.granted === 0) {
        grantedText = intl.formatMessage({id: 'AllBadgesRow.notGranted', defaultMessage: 'Not yet granted.'});
    } else if (badge.multiple) {
        grantedText = intl.formatMessage({id: 'AllBadgesRow.grantedMultiple', defaultMessage: 'Granted {grantedTimes} to {granted} users.'}, {grantedTimes: badge.granted_times, granted: badge.granted});
    } else {
        grantedText = intl.formatMessage({id: 'AllBadgesRow.granted', defaultMessage: 'Granted to {granted} users.'}, {granted: badge.granted});
    }

    return (
        <div className='AllBadgesRow'>
            <a
                className='badge-icon'
                onClick={() => onClick(badge)}
            >
                <span>
                    <BadgeImage
                        badge={badge}
                        size={32}
                    />
                </span>
            </a>
            <div>
                <div className='badge-name'>{badge.name}</div>
                <div className='badge-description'>{markdown(badge.description)}</div>
                <div className='badge-type'>{intl.formatMessage({id: 'AllBadgesRow.type', defaultMessage: 'Type: {typeName}'}, {typeName: badge.type_name})}</div>
                <div className='granted-by'>{grantedText}</div>
            </div>
        </div>
    );
};

export default AllBadgesRow;
