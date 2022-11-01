import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import classnames from 'classnames';
import CloseButton from 'components/CloseButton';
import Button from 'components/Button';

import './modal.scss';

const Modal = ({ title, children, onClose, className, height = 380, onDone, doneTitle = "Done", disableDone }) => {
    const [portalContainer, setPortalContainer] = useState(null);

    useEffect(() => {
        const container = document.querySelector("main");

        if (!container) {
            return;
        }

        setPortalContainer(container);
    }, []);

    if (!portalContainer) {
        return null;
    }

    return ReactDOM.createPortal(
        <div className="modal-outer-wrapper">
            <div className={classnames("modal-inner-wrapper", className)} style={{ height: `${height}px` }}>
                <div className="modal-title">{title}</div>
                <div className="modal-content">{children}</div>
                <CloseButton onClose={onClose} />
                <div className="modal-actions">
                    <Button secondary onClick={onClose}>Cancel</Button>
                    <Button onClick={onDone} disabled={disableDone}>{doneTitle}</Button>
                </div>
            </div>
        </div>,
        portalContainer
    );
}

export default Modal;