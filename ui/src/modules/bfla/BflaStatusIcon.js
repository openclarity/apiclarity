import React from 'react';
import Icon, { ICON_NAMES } from 'components/Icon';
import Tooltip from 'components/Tooltip';

import COLORS from 'utils/scss_variables.module.scss';

export const BFLA_STATUS_TYPES_MAP = {
    SUSPICIOUS_MEDIUM: {
        value: "SUSPICIOUS_MEDIUM",
        label: "Denied",
        icon: ICON_NAMES.ALERT,
        tooltip: "Suspicious Source Denied",
        color: COLORS["color-warning"]
    },
    SUSPICIOUS_HIGH: {
        value: "SUSPICIOUS_HIGH",
        label: "Allowed",
        icon: ICON_NAMES.ALERT,
        tooltip: "Suspicious Source Allowed",
        color: COLORS["color-error"]
    },
    LEARNING: {
        value: "LEARNING",
        label: "Learning",
        icon: ICON_NAMES.REFRESH,
        color: COLORS["color-risk-unknown"]
    },
    NO_SPEC: {
        value: "NO_SPEC",
        label: "Missing provided and reconstructed spec",
        icon: ICON_NAMES.REFRESH,
        color: COLORS["color-risk-unknown"]
    },
    LEGITIMATE: {
        value: "LEGITIMATE",
        icon: ICON_NAMES.CHECK_MARK,
        color: COLORS["color-success"]
    }
};

const BflaStatusIcon = ({id, bflaStatusType, onClick}) => {
    const tooltipId = `bfla-status-${id}`;
    const {icon, tooltip, color} = BFLA_STATUS_TYPES_MAP[bflaStatusType] || BFLA_STATUS_TYPES_MAP['NO_STATUS'];

    return (
        <div className="bfla-status-icon" style={{ width: "22px" }} onClick={onClick}>
                <React.Fragment>
                    <div data-tip data-for={tooltipId}><Icon name={icon} style={{ color }} /></div>
                    {tooltip && <Tooltip id={tooltipId} text={tooltip} props />}
                </React.Fragment>
        </div>
    );
};

export default BflaStatusIcon;
