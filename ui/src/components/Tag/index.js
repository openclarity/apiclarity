import React from 'react';
import classnames from 'classnames';

import './tag.scss';

const Tag = ({children, rounded}) => (
    <div className="tag-wrapper"><div className={classnames("tag-container", {rounded})}>{children}</div></div>
)

export default Tag;