import React from 'react';
import classnames from 'classnames';

import './page-container.scss';

const PageContainer = ({children, className}) => (
    <div className={classnames("page-container", className)}>
        {children}
    </div>
);

export default PageContainer;