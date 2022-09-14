import React from 'react';
import classnames from 'classnames';
import RoundIconContainer from 'components/RoundIconContainer';
import { ICON_NAMES } from 'components/Icon';
import Tooltip, { TOOLTIP_PLACEMENTS } from 'components/Tooltip';

import './info-icon.scss';

const Icon = ({onClick, className}) => (
    <RoundIconContainer
        name={ICON_NAMES.INFO}
        className={classnames("info-icon-container", {[className]: className})}
        onClick={onClick}
    />
);

const InfoIcon = ({tooltipId, text, onClick, placement=TOOLTIP_PLACEMENTS.RIGHT}) => {
    if (!!tooltipId) {
        return (
            <React.Fragment>
                <div className="info-icon-tooltip-wrapper" data-tip data-for={tooltipId}>
                    <Icon className />
                </div>
                <Tooltip id={tooltipId} text={text} placement={placement} />
            </React.Fragment>
        );
    }

    return <Icon className={classnames({"clickable": !!onClick}, "info-icon-non-tooltip")} onClick={onClick} />
}

export default InfoIcon;
