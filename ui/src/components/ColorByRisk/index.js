import React from 'react';
import classnames from 'classnames';
import { API_RISK_ITEMS } from 'utils/systemConsts';

const ColorByRisk = ({tag : Tag='div', risk, withLabel=false, children, labelSuffix="", isText=true})  => {
    const formattedRisk = risk || API_RISK_ITEMS.UNKNOWN.value;
    const riskClassName = classnames("global-risk-color", {text: isText}, formattedRisk.toLowerCase());

    return (
        <Tag className={riskClassName}>
            {withLabel ? `${API_RISK_ITEMS[formattedRisk].label} ${labelSuffix}` : children}
        </Tag>
    )
}

export default ColorByRisk;
