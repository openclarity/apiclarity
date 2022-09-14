import React from 'react';
import ReactTooltip from 'react-tooltip';

import './tooltip.scss';

export const TOOLTIP_PLACEMENTS = {
    TOP: "top",
    BOTTOM: "bottom",
    RIGHT: "right",
    LEFT: "left"
}

const Tooltip = ({id, text}) => (
    <ReactTooltip
        id={id}
        className="ac-tooltip"
        effect='solid'
        place="top"
        textColor="white"
        backgroundColor="rgba(34, 37, 41, 0.8)"
    >
        <span>{text}</span>
    </ReactTooltip>
)

export const TooltipWrapper = ({id, text, placement, center, children}) => (
    <React.Fragment>
        <div data-tip data-for={id} style={{display: "inline-block"}}>{children}</div>
        <Tooltip id={id} text={text} placement={placement} center={center} />
    </React.Fragment>
)

export default Tooltip;
