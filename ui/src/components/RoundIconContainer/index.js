import React from 'react';
import classnames from 'classnames';
import Icon from 'components/Icon';

import './round-icon-container.scss';

const RoundIconContainer = ({name, onClick, className, small=false, disabled}) => {
    const containerClassName = classnames(
        "round-icon-container",
        {clickable: !!onClick},
        {[className]: className, small},
        {disabled: disabled}
    );

    return (
        <div className={containerClassName} onClick={event => disabled || !onClick ? undefined : onClick(event)}>
            <Icon name={name} />
        </div>
    );
};

export default RoundIconContainer;
