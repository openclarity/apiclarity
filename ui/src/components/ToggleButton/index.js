import React from 'react';
import Toggle from 'react-toggle-button';
import COLORS from 'utils/scss_variables.module.scss';

import './toggle-button.scss';

const ToggleButton = ({ title, value, onChange }) => (
    <label className="toggle-button-wrapper">
        <Toggle
            inactiveLabel=""
            activeLabel=""
            colors={{
                activeThumb: {
                    base: "white",
                    border: COLORS["color-main-light"]
                },
                inactiveThumb: {
                    base: COLORS["color-grey"],
                },
                active: {
                    base: COLORS["color-main-light"]
                },
                inactive: {
                    base: COLORS["color-grey-light"]
                }
            }}
            trackStyle={{
                height: "12px",
                width: "40px",
                border: "none"
            }}
            thumbStyle={{
                height: "22px",
                width: "22px",
                border: value ? `1px solid ${COLORS["color-main-light"]}` : "none"
            }}
            thumbAnimateRange={[0, 18]}
            value={value}
            onToggle={() => onChange(!value)}
        />
        <div className="toggle-button-title">{title}</div>
    </label>
)

export default ToggleButton;