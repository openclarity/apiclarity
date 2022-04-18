import React, {useEffect} from 'react';
import Modal from 'components/Modal';
import { useFetch, FETCH_METHODS } from 'hooks';
import Loader from 'components/Loader';

const BflaModal = ({eventId, type, onClose, onSuccess}) => {
    const [{loading: updatePending, data: updateData }, updateBflaWarning] = useFetch(`modules/bfla/event`, {loadOnMount: false});

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

    const titleType = type === 'approve' ? 'legitimate' : 'illegitimate';
    const loading = updatePending;

    return (
        <Modal
            title={`Mark the API call as ${titleType}`}
            height={230}
            onClose={onClose}
            doneTitle="Continue"
            onDone={fetchModelAndUpdate}
        >
            <div>Would you like to mark the selected API call as {titleType}? This will update the API authorization model accordingly.</div>
            {loading && <Loader />}
        </Modal>
    );
};

export default BflaModal;
