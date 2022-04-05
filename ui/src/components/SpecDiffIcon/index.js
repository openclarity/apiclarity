import React from 'react';
import Icon, { ICON_NAMES } from 'components/Icon';
import Tooltip from 'components/Tooltip';

import COLORS from 'utils/scss_variables.module.scss';

export const SPEC_DIFF_TYPES_MAP = {
    ZOMBIE_DIFF: {
        value: "ZOMBIE_DIFF",
        label: "Zombie",
        icon: ICON_NAMES.ZOMBIE,
        tooltip: "Zombie: a deprecated API has been detected",
        color: COLORS["color-error"]
    },
    SHADOW_DIFF: {
        value: "SHADOW_DIFF",
        label: "Shadow",
        icon: ICON_NAMES.SHADOW,
        tooltip: "Shadow: an undocumented API has been detected",
        color: COLORS["color-error"]
    },
    GENERAL_DIFF: {
        value: "GENERAL_DIFF",
        label: "General diff",
        icon: ICON_NAMES.ALERT,
        tooltip: "General diff",
        color: COLORS["color-warning"]
    },
    NO_DIFF: {
        value: "NO_DIFF",
        label: "No diff"
    }
}

const SpecDiffIcon = ({id, specDiffType}) => {
    const tooltipId = `spec-diff-${id}`;
    const {icon, tooltip, color} = SPEC_DIFF_TYPES_MAP[specDiffType] || {};

    if (!icon) {
        return null;
    }

    return (
        <div className="spec-diff-icon" style={{width: "22px"}}>
            <div data-tip data-for={tooltipId}><Icon name={icon} style={{color}} /></div>
            <Tooltip id={tooltipId} text={tooltip} />
        </div>
    )
}

export default SpecDiffIcon;