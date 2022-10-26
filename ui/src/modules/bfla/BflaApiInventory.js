import React, { useEffect, useState } from 'react';
import { useFetch } from 'hooks';
import { has, isEmpty } from 'lodash';
import Loader from 'components/Loader';
import { MODULE_TYPES } from '../MODULE_TYPES.js';
import NoSpecScreen from './NoSpecsScreen';
import DataCollectionScreen from './DataCollectionScreen';
import DataCollectionInProgressScreen from './DataCollectionInProgressScreen';
import DataCollectedScreen from './DataCollectedScreen';

import BFLA_ACTIONS from './actions';
import BFLA_UTILS from './utils.js';

import './bfla.scss';

const SPEC_TYPE = {
    NONE: 'NONE',
    RECONSTRUCTED: 'RECONSTRUCTED',
    PROVIDED: 'PROVIDED'
}

const BflaApiInventory = (props) => {
    const { id: apiId } = props;
    const authModelURL = `modules/bfla/authorizationModel/${apiId}`;
    const stateURL = `modules/bfla/authorizationModel/${apiId}/state`;
    const [{ loading, data }, updateAuthModel] = useFetch(authModelURL);
    const [{ loading: isLoadingCheck, data: checkStateData }, checkStateBFLA] = useFetch(stateURL);
    const [isLoading, setIsLoading] = useState(false);
    const [bflaTabStatus, setBflaTabStatus] = useState(BFLA_UTILS.BFLA_TAB_STATUS.NO_SPEC);


    const handleStartModelLearning = () => {
        setIsLoading(true)
        BFLA_ACTIONS.putBflaApiAuthModelStartLearning(apiId)
            .then((data) => {
                setIsLoading(false)
                setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.LEARNING)
            })
            .catch(err => {
                setIsLoading(false)
                setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.DATA_COLLECTION)
            })
    }

    const handleStopModelLearning = () => {
        setIsLoading(true)
        BFLA_ACTIONS.putBflaApiAuthModelStopLearning(apiId)
            .then((data) => {
                setIsLoading(false)
                updateAuthModel()
                checkStateBFLA()
            })
            .catch(err => {
                setIsLoading(false)
                setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.DATA_COLLECTION)
            })
    }

    const handleReset = () => {
        setIsLoading(true)
        BFLA_ACTIONS.postBflaApiAuthModelReset(apiId).then(({ data }) => {
            setIsLoading(false)
            setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.DATA_COLLECTION)
        }).catch(err => {
            setIsLoading(false)
            setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.DATA_COLLECTION)
        })
    }

    const handleMarkAsLegitimate = (method, path, k8sClientUid) => {
        setIsLoading(true)
        BFLA_ACTIONS.putBflaApiAuthModelApprove(apiId, method, path, k8sClientUid).then(({ data }) => {
            setIsLoading(false)
            updateAuthModel()
        }).catch(err => {
            setIsLoading(false)
        })
    }

    const handleMarkAsIlegitimate = (method, path, k8sClientUid) => {
        setIsLoading(true)
        BFLA_ACTIONS.putBflaApiAuthModelDeny(apiId, method, path, k8sClientUid).then(({ data }) => {
            setIsLoading(false)
            updateAuthModel()
        }).catch(err => {
            setIsLoading(false)
        })
    }

    const handleStartDetection = () => {
        setIsLoading(true)
        BFLA_ACTIONS.putBflaApiAuthModelStartDetection(apiId)
            .then((data) => {
                setIsLoading(false)
                checkStateBFLA()
            })
            .catch(err => {
                setIsLoading(false)
            })
    }

    const handleStopModelDetecting = () => {
        setIsLoading(true)
        BFLA_ACTIONS.putBflaApiAuthModelStopDetection(apiId)
            .then((data) => {
                setIsLoading(false)
                updateAuthModel()
                checkStateBFLA()
            })
            .catch(err => {
                setIsLoading(false)
                setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.DATA_COLLECTION)
            })
    }

    useEffect(() => {
        const isLearningOrDetecting = checkStateData === BFLA_UTILS.BFLAState.BFLA_DETECTING || checkStateData === BFLA_UTILS.BFLAState.BFLA_LEARNING

        if (isLearningOrDetecting) {
            if (checkStateData === BFLA_UTILS.BFLAState.BFLA_DETECTING) {
                setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.DETECTING)
            } else {
                setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.LEARNING)
            }
        }

        if (has(data, "specType") && data.specType === SPEC_TYPE.NONE) {
            setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.NO_SPEC)
        }

        if (has(data, "specType") &&
            (
                data.specType === SPEC_TYPE.RECONSTRUCTED ||
                data.specType === SPEC_TYPE.PROVIDED
            ) &&
            isEmpty(data.operations)
        ) {
            setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.DATA_COLLECTION)
        }

        if (has(data, 'operations') && !isEmpty(data.operations) && !isLearningOrDetecting) {
            setBflaTabStatus(BFLA_UTILS.BFLA_TAB_STATUS.DATA_AVAILABLE)
        }
    }, [data, checkStateData])

    if (loading || isLoading || isLoadingCheck) {
        return <Loader />;
    }

    switch (bflaTabStatus) {
        case BFLA_UTILS.BFLA_TAB_STATUS.NO_SPEC:
            return <NoSpecScreen id={apiId} />
        case BFLA_UTILS.BFLA_TAB_STATUS.DATA_COLLECTION:
            return <DataCollectionScreen
                handleStartModelLearning={handleStartModelLearning}
                id={apiId}
            />
        case BFLA_UTILS.BFLA_TAB_STATUS.DETECTING:
            return <DataCollectionInProgressScreen
                isLearning={false}
                handleStop={handleStopModelDetecting}
            />
        case BFLA_UTILS.BFLA_TAB_STATUS.LEARNING:
            return <DataCollectionInProgressScreen
                isLearning={true}
                handleStop={handleStopModelLearning}
            />
        case BFLA_UTILS.BFLA_TAB_STATUS.DATA_AVAILABLE:
            return <DataCollectedScreen
                data={data}
                handleReset={handleReset}
                handleMarkAsLegitimate={handleMarkAsLegitimate}
                handleMarkAsIlegitimate={handleMarkAsIlegitimate}
                handleStartDetection={handleStartDetection}
                handleStartLearning={handleStartModelLearning}
            />
        default:
            return <></>;

    }
};

const bflaApiInventory = {
    name: 'BFLA',
    component: BflaApiInventory,
    endpoint: '/bfla',
    type: MODULE_TYPES.INVENTORY_DETAILS
};

export default bflaApiInventory;
