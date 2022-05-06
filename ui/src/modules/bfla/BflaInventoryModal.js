import React, {useEffect} from 'react';
import Modal from 'components/Modal';
import { useFetch, FETCH_METHODS } from 'hooks';
import Loader from 'components/Loader';

const AUTH_MODEL_OPERATIONS = {
    APPROVE: 'approve',
    DENY: 'deny'
};

const BflaInventoryModal = ({client, method, path, url, onClose, onSuccess}) => {
    const [{loading: updatePending, data: updateData }, updateAuthModel] = useFetch(url, {loadOnMount: false});
    const {authorized, k8s_object} = client || {};
    const {name, uid} = k8s_object || {};

    useEffect(() => {
        if (updateData) {
            onClose();
            onSuccess();
        }
    }, [updateData, onSuccess, onClose]);

    const operation = authorized ? AUTH_MODEL_OPERATIONS.DENY : AUTH_MODEL_OPERATIONS.APPROVE;
    const updateModel = () => updateAuthModel({
        formatUrl: (url) => `${url}/${operation}`,
        queryParams: { method: method, path: path, k8sClientUid: uid },
        method: FETCH_METHODS.PUT
    });

    const titleType = authorized ?'unauthorize' : 'authorize';
    const loading = updatePending;

    return (
        <Modal
            title={`${titleType} client`}
            height={230}
            onClose={onClose}
            doneTitle="Continue"
            onDone={updateModel}
        >
            <div>Would you like to {titleType} client "{name}" for this API call?</div>
            {loading && <Loader />}
        </Modal>
    );
};

export default BflaInventoryModal;
