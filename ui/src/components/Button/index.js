import React from 'react';
import classnames from 'classnames';

import './button.scss';

const Button = ({children, secondary, disabled, onClick, className}) => (
    <button
        className={classnames("ag-button", className, {"ag-button--primary": !secondary}, {"ag-button--secondary": secondary}, {clickable: !!onClick})}
        disabled={disabled}
        onClick={event => !disabled && onClick ? onClick(event) : null}
    >
        {children}
    </button>
)

export default Button;