import React, { useState } from 'react';
import classnames from 'classnames';
import Arrow from 'components/Arrow';

import './accordion.scss';

const Accordion = ({ title, customTitle, children, className, isEmpty = false }) => {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <div className={classnames("accordion-wrapper", { [className]: !!className })}>
            <div className={classnames("accordion-header", { "is-empty": isEmpty })} onClick={!isEmpty ? () => setIsOpen(isOpen => !isOpen) : null}>
                {customTitle ? customTitle : <div className="accordion-title">{title}</div>}
                {!isEmpty && <Arrow name={isOpen ? "top" : "bottom"} small />}
            </div>
            {(isOpen && !isEmpty) &&
                <div className="accordion-content">
                    {children}
                </div>
            }
        </div>
    )
}

export default Accordion;
