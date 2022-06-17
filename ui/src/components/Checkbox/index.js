import React from 'react';
import classnames from 'classnames';

import './checkbox.scss';

const Checkbox = ({checked, name, title, onChange, className, halfSelected, disabled}) => (
    <div className="ag-checkbox-wrapper">
        <label className={classnames("ag-checkbox", className, {disabled})}>
            <input type="checkbox" value={checked} name={name} checked={checked} onChange={(event) => disabled ? null : onChange(event)} />
            <span className={classnames("checkmark", {"half-selected": halfSelected})}></span>
            <div className="ag-checkbox-title">{title}</div>
        </label>
    </div>
);

export default Checkbox;
