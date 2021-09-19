import React from 'react';
import ReactTooltip from 'react-tooltip';

import './tooltip.scss';

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

export default Tooltip;

