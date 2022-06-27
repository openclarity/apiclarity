import React from 'react';
import Tag from 'components/Tag';

import COLORS from 'utils/scss_variables.module.scss';

const COLOR_MAPPING = {
    POST: COLORS["color-success"],
    GET: COLORS["color-main-light"],
    PUT: COLORS["color-warning"],
    PATCH: COLORS["color-status-blue"],
    OPTIONS: COLORS["color-main"],
    HEAD: COLORS["color-status-violet"],
    DELETE: COLORS["color-error"]
}

const MethodTag = ({method}) => (
    <Tag color={COLOR_MAPPING[method]}>{method}</Tag>
)

export default MethodTag;