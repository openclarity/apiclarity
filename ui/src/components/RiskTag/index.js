import React from 'react';
import classnames from 'classnames';
import './risk-tag.scss';

const ALERT_RISKS = {
    INFO: {value: "INFO", label: "Info"},
    WARN: {value: "WARN", label: "Warn"}
}

const RiskTag = ({risk, label}) => {
    const formattedRisk = risk || ALERT_RISKS.INFO.value;

    return (
        <div>
            <div className={classnames("risk-tag-wrapper", formattedRisk.toLowerCase())}>
                {label || ALERT_RISKS[formattedRisk].label}
            </div>
        </div>
    )
}

export default RiskTag;
