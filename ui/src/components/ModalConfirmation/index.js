import React from 'react';
import classnames from 'classnames';
import Modal from 'components/SideDrawer';
// import Text, { TEXT_TYPES } from 'components/Text';
import Button from 'components/Button';

import './modal-confirmation.scss';

const ModalConfirmation = ({title, message, confirmTitle, onCancle, onConfirm, loading, className, confirmAlert=false}) => (
    <Modal className={classnames("modal-confirmation", className)} onClose={onCancle} center={true} loading={loading}>
        <div className="confirmation-title">{title}</div>
        <div>{message}</div>
        <div className="confirmation-actions-container">
            <Button className="confirmation-cancel-button" tertiary onClick={onCancle}>Cancel</Button>
            <Button className={classnames("confirmation-confirmation-button", {alert: confirmAlert})} onClick={onConfirm}>{confirmTitle}</Button>
        </div>
    </Modal>
);

export default ModalConfirmation;
