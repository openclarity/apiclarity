import React from 'react';
import Button from 'components/Button';

import './message-text-button.scss';

const MessageTextButton = ({children, onClick, disabled}) => (
    <Button tertiary className="message-text-button" onClick={onClick} disabled={disabled}>
        {children}
    </Button>
)

export default MessageTextButton;
