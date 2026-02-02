import React from 'react';
import {useSelector} from 'react-redux';
import {IntlProvider} from 'react-intl';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/users';

import {getMessages} from '../i18n';

interface Props {
    children: React.ReactNode;
}

const IntlProviderWrapper: React.FC<Props> = ({children}) => {
    const currentUser = useSelector(getCurrentUser);

    const mattermostLocale = currentUser?.locale || 'ko';
    const language = mattermostLocale.split('-')[0];

    return (
        <IntlProvider
            locale={language}
            messages={getMessages(language)}
        >
            {children}
        </IntlProvider>
    );
};

export default IntlProviderWrapper;
