import React from 'react';
import { useFetch } from 'hooks';
import Loader from 'components/Loader';
import { MODULE_TYPES } from '../MODULE_TYPES.js';
import BflaInventoryModal from './BflaInventoryModal';
import emptySelectImage from 'utils/images/select.svg';

import collectionProgress from './collection.svg';

import NoSpecScreen from './NoSpecsScreen';
import DataCollectionScreen from './DataCollectionScreen';
import DataCollectionInProgressScreen from './DataCollectionInProgressScreen';
import StartDetectionResumeLearningScreen from './StartDetectionResumeLearningScreen';
import DataCollectedScreen from './DataCollectedScreen';

import BFLA_UTILS from './utils.js';

import './bfla.scss';

const mockData = {
    tags: [
        {
            name: 'tag1',
            paths: [
                {
                    path: '/PATH1',
                    method: 'POST',
                    isLegitimate: true,
                    clients: [
                        {
                            name: 'PATH1 client 1',
                            lastObserved: new Date(),
                            lastStatusCode: 200,
                            isLegitimate: true,
                            principles: [
                                {
                                    principleType: 'principleType',
                                    name: 'name1',
                                    ip: '192.168.0.1'
                                }
                            ]
                        }
                    ]
                },
                {
                    path: '/PATH2',
                    method: 'GET',
                    isLegitimate: false,
                    clients: [
                        {
                            name: 'PATH2 client 2',
                            lastObserved: new Date(),
                            lastStatusCode: 401,
                            isLegitimate: false,
                        }
                    ]
                }
            ],
            isLegitimate: true,
        },
        {
            name: 'tag2',
            paths: [

            ],
            isLegitimate: false,
        },
    ]
}


const BflaApiInventory = (props) => {
    const { id: apiId } = props;
    const authModelURL = `modules/bfla/authorizationModel/${apiId}`;
    const [{ loading, data }, updateAuthModel] = useFetch(authModelURL);

    const { bflaStatus } = { bflaStatus: BFLA_UTILS.BFLAStatus.LEARNING };

    if (loading) {
        return <Loader />;
    }


    switch (bflaStatus) {
        case BFLA_UTILS.BFLAStatus.NO_SPEC:
            return <NoSpecScreen id={apiId} />
        case BFLA_UTILS.BFLAStatus.LEARNING:
            return <DataCollectionInProgressScreen id={apiId} />
        default:
            return <DataCollectedScreen />
    }

};

const bflaApiInventory = {
    name: 'BFLA',
    component: BflaApiInventory,
    endpoint: '/bfla',
    type: MODULE_TYPES.INVENTORY_DETAILS
};

export default bflaApiInventory;
