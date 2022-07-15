import React from 'react';
import classnames from 'classnames';

import COLORS from 'utils/scss_variables.module.scss';
import './tag.scss';

const Tag = ({children, rounded, color=COLORS["color-main"]}) => (
    <div className="tag-wrapper"><div style={{backgroundColor: color}} className={classnames("tag-container", {rounded})}>{children}</div></div>
)

export default Tag;