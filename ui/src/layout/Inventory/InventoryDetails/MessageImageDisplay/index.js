import React from 'react';
import classnames from 'classnames';

import './message-image-display.scss';

const MessageImageDisplay = ({message, subMessage, image, className}) => (
    <div className={classnames("message-image-display-wrapper", className)}>
        <div className="message-image-display-message">{message}</div>
        {!!subMessage && <div className="message-image-display-sub-message">{subMessage}</div>}
        <img src={image} alt="message" />
    </div>
)

export default MessageImageDisplay;
