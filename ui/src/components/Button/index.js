import React from 'react';
import classnames from 'classnames';
import Icon from 'components/Icon';

import './button.scss';

const Button = ({ type = "button", className, children, onClick, disabled = false, secondary = false, tertiary = false, icon }) => (
    <button
        type={type}
        className={classnames(
            "ag-button",
            { "ag-button--primary": !secondary && !tertiary },
            { "ag-button--secondary": secondary },
            { "ag-button--tertiary": tertiary },
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