import React from 'react';
import Modal from 'components/Modal';
import Icon, { ICON_NAMES } from 'components/Icon';

import './confirmation-modal.scss';

const BoldText = ({children}) => <span style={{fontWeight: "bold"}}>{children}</span>;

const ConfirmationModal = ({onClose, onConfirm, inventoryName, pathsCount}) => {
    return (
        <Modal
            title="Approve review"
            onClose={onClose}
            className="review-confirmation-modal"
            height={230}
            onDone={onConfirm} 
            doneTitle="Yes"
        >
            <div>Do you want to create a reconstructed spec for <BoldText>{inventoryName}</BoldText> with the <BoldText>{pathsCount}</BoldText> selected entries?</div>
            <div className="approve-alert">
                <Icon name={ICON_NAMES.ALERT_ROUND} />
                <div>Once approved, it won't be possible to edit or review.</div>
            </div>
        </Modal>
    )
}

export default ConfirmationModal;