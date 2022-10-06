import React from 'react';
import classnames from 'classnames';

export const ModalSection = ({title, className, children}) => (
    <div className={classnames('modal-details-section', className)}>
        <div className="modal-details-section-title">{title}</div>
        <div className="modal-details-section-content">
            {children}
        </div>
    </div>
);

export const ModalTitleDataItem = ({title, className, children}) => (
    <div className={classnames('modal-title-data-item', className)}>
        <div className="modal-title-data-item-title">{title}</div>
        <div className="modal-title-data-item-data">
            {children}
        </div>
    </div>
);
