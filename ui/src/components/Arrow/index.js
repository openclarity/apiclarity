import React from 'react';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';

import './arrow.scss';

export const ARROW_NAMES = {
    TOP: "top",
    BOTTOM: "bottom",
    RIGHT: "right",
    LEFT: "left"
}

const Arrow = ({ name = ARROW_NAMES.TOP, onClick, disabled, small = false, className }) => {
    if (!Object.values(ARROW_NAMES).includes(name)) {
        console.error(`Arrow name '${name}' does not exist`);
    }

    return (
        <Icon
            name={small ? ICON_NAMES.SMALL_ARROW_LEFT : ICON_NAMES.ARROW_LEFT}
            className={classnames("arrow-icon", `${name}-arrow`, { small }, { [className]: !!className }, { clickable: !!onClick })}
            onClick={onClick}
            disabled={disabled}
        />
    );
}

export default Arrow;