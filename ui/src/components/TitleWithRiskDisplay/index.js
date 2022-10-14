import React from 'react';
import classnames from 'classnames';
import RiskTag from 'components/RiskTag';

import './title-with-risk-display.scss';

const TitleWithRiskDisplay = ({title, risk, className, hideRisk=false, customRiskDisplay: CustomRiskDisplay, alertDisplay: AlertDisplay}) => (
    <div className={classnames("title-with-risk-display", className)}>
        <div className="title-with-risk-display-title">{title}</div>
        <div className="title-with-risk-display-risk-wrapper">
            {!!AlertDisplay && <div className="title-with-risk-display-alert"><AlertDisplay /></div>}
            {!hideRisk && <div className="title-with-risk-display-risk">{!!CustomRiskDisplay ? <CustomRiskDisplay /> : <RiskTag risk={risk} />}</div>}
        </div>
    </div>
)

export default TitleWithRiskDisplay;
