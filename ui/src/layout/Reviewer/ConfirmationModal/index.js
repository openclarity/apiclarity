import React, {useState} from 'react';
import Modal from 'components/Modal';
import Icon, { ICON_NAMES } from 'components/Icon';
import DropdownSelect from 'components/DropdownSelect';

import './confirmation-modal.scss';

const OAS_VERSIONS = {
    OASV2: {
        value: 'OASv2.0',
        label: 'OAS V2'
    },
    OASV3: {
        value: 'OASv3.0',
        label: 'OAS V3'
    }
};
const BoldText = ({children}) => <span style={{fontWeight: "bold"}}>{children}</span>;

const ConfirmationModal = ({onClose, onConfirm, inventoryName, pathsCount}) => {
    const [OASVersion, setOASVersion] = useState();
    const oasVersions = Object.values(OAS_VERSIONS);

    return (
        <Modal
            title="Approve review"
            onClose={onClose}
            className="review-confirmation-modal"
            height={355}
            onDone={() => onConfirm(OASVersion.value)}
            disableDone={!OASVersion}
            doneTitle="Yes"
        >
            <div>Do you want to create a reconstructed spec for <BoldText>{inventoryName}</BoldText> with the <BoldText>{pathsCount}</BoldText> selected entries?</div>
            <div className="approve-alert">
                <Icon name={ICON_NAMES.ALERT_ROUND} />
                <div>Once approved, it won't be possible to edit or review.</div>
            </div>
            <div>Choose the OpenAPI specification version of the generated spec.</div>
            <div className="dropdown-title"><BoldText>OAS version</BoldText></div>
            <DropdownSelect
                items={oasVersions}
                onChange={(item) => setOASVersion(item)}
                value={OASVersion}
            />
        </Modal>
    )
}

export default ConfirmationModal;
