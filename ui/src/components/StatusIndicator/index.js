import React from 'react';
import classnames from 'classnames';

import './status-indicator.scss';

const StatusIndicator = ({title, isError}) => (
    <div className="status-indicator-wrapper">
        <div className={classnames("status-indicator", {error: isError})}></div>
        <div className="status-value">{title}</div>
    </div>
);

export default StatusIndicator;