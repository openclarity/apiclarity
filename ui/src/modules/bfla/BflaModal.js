import React, { useEffect } from 'react';
import { useFetch, FETCH_METHODS } from 'hooks';
import Loader from 'components/Loader';
import ModalConfirmation from 'components/ModalConfirmation';

export const MODAL_ACTION_TYPE = {
    APPROVE: 'approve',
    DENY: 'deny'
};

const MODAL_TITLE_TYPE = {
    LEGITIMATE: 'legitimate',
    ILLEGITIMATE: 'illegitimate'
};

const BflaModal = ({ eventId, type, onClose, onSuccess }) => {
    const [{ loading: updatePending, data: updateData }, updateBflaWarning] = useFetch(`modules/bfla/event`, { loadOnMount: false });

    useEffect(() => {
        if (updateData) {
            onClose();
            onSuccess();
        }
    }, [updateData, onSuccess, onClose]);

    const fetchModelAndUpdate = () => updateBflaWarning({
        formatUrl: (url) => `${url}/${eventId}/${type}`,
        method: FETCH_METHODS.PUT
    });

    const titleType = type === MODAL_ACTION_TYPE.APPROVE ? MODAL_TITLE_TYPE.LEGITIMATE : MODAL_TITLE_TYPE.ILLEGITIMATE;

    return (
        <ModalConfirmation
            title={`Mark the API call as ${titleType}`}
            message={`Would you like to mark the selected API call as ${titleType}? This will update the API authorization model accordingly.`}
            confirmTitle="Continue"
            onCancle={onClose}
            onConfirm={fetchModelAndUpdate}
        />
    );
};

export default BflaModal;
