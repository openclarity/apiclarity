import React from 'react';
import classnames from 'classnames';
import Icon from 'components/Icon';

import './icon-with-title.scss';

const IconWithTitle = (props) => {
    const {title, onClick, className, ...iconProps} = props;
    const {disabled, name} = props;

    return (
        <div
            className={classnames("icon-container", `icon-container-${name}`, {disabled}, {[className]: className}, {clickable: !!onClick && !disabled})}
            onClick={event => !disabled && onClick ? onClick(event) : null}
        >
            <Icon {...iconProps} /><div className="icon-title">{title}</div>
        </div>
    );
}

export default IconWithTitle;
