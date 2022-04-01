import React from 'react';
import classnames from 'classnames';
import Icon from 'components/Icon';

import './icon-with-title.scss';

const IconWithTitle = (props) => {
    const {title, onClick, ...iconProps} = props;
    const {name} = props;

    return (
        <div
            className={classnames("icon-container", `icon-container-${name}`, "clickable")}
            onClick={event =>  onClick(event)}
        >
            <Icon {...iconProps}/><div className="icon-title">{title}</div>
        </div>
    );
};

export default IconWithTitle;
