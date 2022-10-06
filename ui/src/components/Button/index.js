import React from 'react';
import classnames from 'classnames';
import Icon from 'components/Icon';

import './button.scss';

const Button = ({ type = "button", className, children, onClick, disabled = false, secondary = false, tertiary = false, icon }) => (
    <button
        type={type}
        className={classnames(
            "scn-button",
            { "scn-button--primary": !secondary && !tertiary },
            { "scn-button--secondary": secondary },
            { "scn-button--tertiary": tertiary },
            className
        )}
        onClick={event => !disabled && onClick ? onClick(event) : null}
        disabled={disabled}
    >
        {!!icon && <Icon name={icon} className="button-icon" />}
        {children}
    </button>
);

export default Button;