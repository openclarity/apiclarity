import React from 'react';
import classnames from 'classnames';

import './title-value-display.scss';

export const TitleValueDisplayRow = ({children}) => (
    <div className="title-value-display-row">{children}</div>
);

const TitleValueDisplay = ({title, children, className}) => (
    <div className={classnames("title-value-display-wrapper", className)}>
        <div className="title-value-display-title">{title}</div>
        <div className="title-value-display-content">{children}</div>
    </div>
);

export default TitleValueDisplay