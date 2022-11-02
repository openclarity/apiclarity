import React from 'react';
import classnames from 'classnames';
import Modal from 'components/Modal';
import Text, { TEXT_TYPES } from 'components/Text';

import './modal-confirmation.scss';

const ModalConfirmation = ({ title, message, confirmTitle, onCancle, onConfirm, loading, className }) => (
    <Modal 
        title={title}
        onClose={onCancle} 
        height
        onDone={onConfirm}
        doneTitle={confirmTitle}
        className={classnames("modal-confirmation", className)} 
        disableDone={loading}
    >
        <Text type={TEXT_TYPES.BODY} withBottomMargin>{message}</Text>
    </Modal>
);

export default ModalConfirmation;
