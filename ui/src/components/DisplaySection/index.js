import React, { useState } from 'react';
import classnames from 'classnames';
import Arrow, { ARROW_NAMES } from 'components/Arrow';

import './display-section.scss';

const DisplaySection = ({className, title, children, hideUnderline, allowClose=false, initialIsClosed=true, noPadding=false}) => {
    const [isClosed, setIsClosed] = useState(initialIsClosed);

    return (
        <div className={classnames("display-section-wrapper", className)}>
            <div onClick={() => setIsClosed(!isClosed)} className={classnames("section-title-wrapper", {"hide-underline": hideUnderline}, {"with-close-toggle": allowClose}, {"with-padding": !noPadding})}>
                <div className="section-title">{title}</div>
                {allowClose && <Arrow name={isClosed ? ARROW_NAMES.BOTTOM : ARROW_NAMES.TOP} small />}
            </div>
            {(!isClosed || !allowClose) && <div className="section-content">{children}</div>}
        </div>
    )
}

export default DisplaySection;
