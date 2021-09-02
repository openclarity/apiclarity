import React from 'react';
import Icon, { ICON_NAMES } from 'components/Icon';

import './close-button.scss';

const CloseButton = ({onClose}) => (
    <Icon name={ICON_NAMES.X_MARK} onClick={onClose} className="close-button" />
)

export default CloseButton;