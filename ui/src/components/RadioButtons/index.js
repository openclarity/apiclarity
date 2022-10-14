import React from 'react';
import classnames from 'classnames';
import InfoIcon from 'components/InfoIcon';

import './radio-buttons.scss';

const RadioButtons = ({name, items, selected, onChange, horizontal=false, small=false, disabled=false}) => (
    <div className={classnames("ps-radio-container", {horizontal}, {small})}>
        {
            items.map(({value, label, tooltip, disabled: disabledItem}) => (
                <React.Fragment key={value}>
                    <label className={classnames("ps-radio-wrapper", {disabled: disabledItem || disabled})}>
                        <div className="ps-radio">
                            <span className="ps-radio-text">{label}</span>
                            <input
                                type="radio"
                                name={name}
                                checked={selected === value}
                                value={value}
                                disabled={disabledItem || disabled}
                                onChange={() => onChange(value)}
                            />
                            <span className="checkmark"></span>
                        </div>
                        {!!tooltip && <InfoIcon tooltipId={`tooltip-${name}-${value}`} text={tooltip} />}
                    </label>
                </React.Fragment>
            ))
        }
    </div>
);

export default RadioButtons;
