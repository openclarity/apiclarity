import React, {useState} from 'react';
import {useFetch} from 'hooks';
import Loader from 'components/Loader';
import Button from 'components/Button';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import BflaStatusIcon, {BFLA_STATUS_TYPES_MAP} from './BflaStatusIcon';
import BflaModal from './BflaModal';
import MODULE_TYPES from '../MODULE_TYPES.js';
import bflaApiInventory from './BflaApiInventory';

import './bfla.scss';

const BflaPlugin = (props) => {
    const {eventId} = props;
    const [{ loading, data: bflaData }, updateBflaEvent] = useFetch(`modules/bfla/event/${eventId}`);
    const [showBflaModal, setShowBflaModal] = useState();

    if (loading) {
        return <Loader />;
    }

    if (!bflaData) {
        return <div>No BFLA data found</div>;
    }

    const {bflaStatus, sourceK8sObject, destinationK8sObject, external} = bflaData;
    const sourceName = external ? 'EXTERNAL' : sourceK8sObject.name;
    const sourceKind = sourceK8sObject ? sourceK8sObject.kind : '';
    const destinationName = destinationK8sObject ? destinationK8sObject.name : '';
    const destinationKind = destinationK8sObject ? destinationK8sObject.kind : '';

    return (
        <div>
            <TitleValueDisplayRow>
                <TitleValueDisplay title={external ? "Source" : "Source Name"}>{sourceName}</TitleValueDisplay>
                {!external &&
                    <TitleValueDisplay title="Source Kind">{sourceKind}</TitleValueDisplay>
                }
                <TitleValueDisplay title="Destination Name">{destinationName}</TitleValueDisplay>
                <TitleValueDisplay title="Destination Kind">{destinationKind}</TitleValueDisplay>
            </TitleValueDisplayRow>
            <TitleValueDisplayRow>
                <BflaStatus bflaStatus={bflaStatus} sourceName={sourceName} />
            </TitleValueDisplayRow>
            <TitleValueDisplayRow>
                {(bflaStatus !== BFLA_STATUS_TYPES_MAP.LEARNING.value && bflaStatus !== BFLA_STATUS_TYPES_MAP.NO_SPEC.value) &&
                    <Button onClick={() => setShowBflaModal(true)} >{bflaStatus === BFLA_STATUS_TYPES_MAP.LEGITIMATE.value ? "Mark as Illegitimate" : "Mark as Legitimate"}</Button>
                }
            </TitleValueDisplayRow>

            {showBflaModal &&
                <BflaModal
                    eventId={eventId}
                    type={bflaStatus === BFLA_STATUS_TYPES_MAP.LEGITIMATE.value ? 'deny' : 'approve' }
                    onClose={() => setShowBflaModal(false)}
                    onSuccess={() => updateBflaEvent()}/>
            }
        </div>
    );
};

const BflaStatus = ({bflaStatus, sourceName}) => {
    const {SUSPICIOUS_HIGH, SUSPICIOUS_MEDIUM, LEGITIMATE, LEARNING, NO_SPEC} = BFLA_STATUS_TYPES_MAP;
    const {value} = BFLA_STATUS_TYPES_MAP[bflaStatus] || {};

    let statusText;
    if (value === LEGITIMATE.value) {
        statusText = <div className="bfla-status-text">This API call seems legitimate.</div>;
    }
    if (value === SUSPICIOUS_HIGH.value) {
        statusText = <div className="bfla-status-text">
                         The pod {sourceName} made this call to the API.
                         This looks suspicious, as it would represent a violation of the current authorization model.  The API server correctly rejected the call
                     </div>;
    }

    if (value === SUSPICIOUS_MEDIUM.value) {
        statusText = <div className="bfla-status-text">
                         The pod {sourceName} made this call to the API. This looks suspicious, as it represents a violation of the current authorization model. Moreover, the API server accepted the call, which implies a possible Broken Function Level Authorisation. Please verify authorisation implementation in the API server.
                     </div>;
    }

    if (value === LEARNING.value) {
        statusText = <div className="bfla-status-text">
                         Data collection in progress.
                     </div>;
    }

    if (value === NO_SPEC.value) {
        statusText = <div className="bfla-status-text">
                         Please either provide a spec or reconstruct one in order to enable BFLA detection for this API.
                     </div>;
    }

    let bflaStatusMesage =
        <TitleValueDisplay
            className="bfla-status"
            title={<div className="bfla-status-title"> <span>Broken Function Level Authorization Alert:</span> <BflaStatusIcon bflaStatusType={bflaStatus} /> </div>}>
                {statusText}
        </TitleValueDisplay>;

    if (value === LEGITIMATE.value || value === LEARNING.value || value === NO_SPEC.value) {
        bflaStatusMesage = <div className="bfla-status bfla-status-title">
                               {statusText}{value !== NO_SPEC.value && <BflaStatusIcon bflaStatusType={bflaStatus} />}
        </div>;
    }

    return (
        bflaStatusMesage
    );
};

const bfla = {
    name: 'BFLA',
    moduleName: 'bfla',
    component: BflaPlugin,
    endpoint: '/bfla',
    type: MODULE_TYPES.EVENT_DETAILS
};

export {
    bfla,
    bflaApiInventory
};
