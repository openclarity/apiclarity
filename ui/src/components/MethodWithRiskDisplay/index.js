import React from 'react';
import TitleWithRiskDisplay from 'components/TitleWithRiskDisplay';

export const MethodWithRiskDisplay = ({path, method, risk, customRiskDisplay, alertDisplay}) => (
    <TitleWithRiskDisplay
        title={<span>{method}<span style={{ fontWeight: "normal", marginLeft: "5px" }}>{path}</span></span>}
        risk={risk}
        customRiskDisplay={customRiskDisplay}
        alertDisplay={alertDisplay}
    />
);

export default MethodWithRiskDisplay;
