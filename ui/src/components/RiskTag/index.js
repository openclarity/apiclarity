import React from 'react';
import classnames from 'classnames';
import { SYSTEM_RISKS } from 'utils/utils';

import './risk-tag.scss';

const RiskTag = ({risk, label}) => {
    const formattedRisk = risk || SYSTEM_RISKS.UNKNOWN.value;

    return (
        <div>
            <div className={classnames("risk-tag-wrapper", formattedRisk.toLowerCase())}>
                {label || SYSTEM_RISKS[formattedRisk].label}
            </div>
        </div>
    )
}

export default RiskTag;
