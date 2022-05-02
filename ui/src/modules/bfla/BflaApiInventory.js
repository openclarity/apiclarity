import React, {useState, useEffect} from 'react';
import {useFetch, usePrevious} from 'hooks';
import Loader from 'components/Loader';
import ListDisplay from 'components/ListDisplay';
import Tag from 'components/Tag';
import Table from 'components/Table';
import Arrow, { ARROW_NAMES } from 'components/Arrow';
import Button from 'components/Button';
import MODULE_TYPES from '../MODULE_TYPES.js';
import BflaInventoryModal from './BflaInventoryModal';
import emptySelectImage from 'utils/images/select.svg';

import collectionProgress from './collection.svg';

import './bfla.scss';

const NotSelected = ({title}) => (
    <div className="not-selected-wrapper">
        <div className="not-selected-title">{title}</div>
        <img src={emptySelectImage} alt="no path selected" />
    </div>
);

const BackHeader = ({title, onBack}) => (
    <div className="selected-back-header">
        <Arrow name={ARROW_NAMES.LEFT} onClick={onBack} />
        <div>{title}</div>
    </div>
);

const MethodTitle = ({method, path}) => (
    <div className="method-item-title"><Tag>{method}</Tag><span>{path}</span></div>
);

const SelectedAuthClientDisplay = ({client, onBack, authorizeClient}) => {
    const {external, k8s_object, end_users} = client || {};
    const name = external ? "EXTERNAL" : k8s_object.name || {};
    const endUsers = end_users || [];

    const columns = [
        {
            Header: 'Type',
            id: 'type',
            Cell: ({row}) => {
                const {source} = row.original;
                return <Tag>{source}</Tag>;
            },
            width: 30
        },
        { Header: 'Name', id: 'name', accessor: 'id', width: 50 },
        { Header: 'IP', id: 'ip', accessor: 'ip_address' }
    ];

    const BackButtonTitle = `authorized clients/${name}`;

    return (
        <React.Fragment>
            <div className="client-action-wrapper">
                <BackHeader title={BackButtonTitle} onBack={onBack} />
                <Button className="button-add" onClick={() => authorizeClient(client)}>MARK AS ILLEGITIMATE</Button>
            </div>
                <div className="authorized-clients">
                    <Table
                        noResultsTitle={"this event"}
                        columns={columns}
                        data={{items: endUsers, total: endUsers.length}}
                        withPagination={false}
                    />
                </div>
        </React.Fragment>
    );
};

const SelectedPathDisplay = ({data, url, onBack, refresh }) => {
    const [selectedClient, setSelectedClient] = useState();
    const [selectedAuthClient, setSelectedAuthClient] = useState();
    const {id, method, path, audience} = data;
    let authorizedClients = [];
    let violatingClients = [];

    const prevId = usePrevious(id);

    useEffect(() => {
        if (id !== prevId) {
            setSelectedAuthClient(null);
        }
    }, [id, prevId, selectedAuthClient]);

    audience.forEach((a, idx) => {
        a.authorized ? authorizedClients.push({ ...a, id: idx }) : violatingClients.push(a);
    });

    const authorizeClient = (client) => {
        setSelectedClient(client);
    };

    const displayTitle = method && path;

    return (
        <div className="tag-selected-wrapper">
            {selectedAuthClient ? <SelectedAuthClientDisplay client={selectedAuthClient} onBack={() => setSelectedAuthClient(null)} authorizeClient={authorizeClient}/> :
                <React.Fragment>
                    <BackHeader title={displayTitle && <MethodTitle method={method} path={path} />} onBack={onBack} />
                    <div className="tag-selected-methods-list">
                        <div className="authorized-clients">
                            <div className="clients-list-title">List of authorized clients</div>
                            <ListDisplay
                                items={authorizedClients}
                                itemDisplay={({ k8s_object, external }) => <div className="client-list-item-title">{
                                    external? "EXTERNAL" :k8s_object.name
                                }</div>}
                                selectedId={!!selectedAuthClient ? selectedAuthClient.id : null}
                                onSelect={(client) => setSelectedAuthClient(client)}
                            />

                        </div>

                        <div className="violating-clients">
                            <div className="clients-list-title">List of violating clients</div>
                            {violatingClients.map((c) => {
                                return <div className="client-list-item-wrapper" key={c.external ? "EXTERNAL": c.k8s_object.name}>
                                    <div className="client-list-item-title">{c.external ? "EXTERNAL": c.k8s_object.name}</div>
                                    <Button className="button-add" onClick={() => authorizeClient(c)}>Authorize</Button>
                                </div>;
                            })}
                        </div>
                    </div>
                </React.Fragment>
            }

            {selectedClient &&
                <BflaInventoryModal
                    url={url}
                    client={selectedClient}
                    method={method}
                    path={path}
                    onClose={() => setSelectedClient(null)}
                    onSuccess={() => refresh()}/>
            }
        </div>
    );
};

const DataCollectionInProgress = ({title}) => (
        <div className="in-progress-wrapper">
            <div className="in-progress-title">{title}</div>
            <img src={collectionProgress} alt="data collection in progress" />
        </div>
);

const BflaTab = ({data, url, loading, refresh}) => {
    const [selectedPathData, setSelectedPathData] = useState(null);
    const {operations} = data || [];
    const methodPathList = operations ? operations.map((x, idx) => ({ id: idx, ...x })) : [];

    return (
        <div className="bfla-tab-wrapper">
            {loading ? <Loader /> :

                <div className="spec-display-wrapper">
                    <div className="select-pane">
                        <ListDisplay
                            items={methodPathList}
                            itemDisplay={({ method, path }) => <MethodTitle method={method} path={path} />}
                            selectedId={!!selectedPathData ? selectedPathData.id : null}
                            onSelect={selected => setSelectedPathData(selected)}
                        />
                    </div>
                    <div className="display-pane">
                        {!selectedPathData ? <NotSelected title="Select a path to see details." /> :
                         <SelectedPathDisplay data={selectedPathData} refresh={refresh} url={url} onBack={() => setSelectedPathData(null)} />}
                    </div>
             </div>
            }
        </div>
    );

};

const collection_in_progress = {
    audience: [
        {
            authorized: true,
            k8s_object: {
                name: 'TBD'
            },
            method: '',
            path: ''
        },
        {
            authorized: false,
            k8s_object: {
                name: 'TBD'
            },
            method: '',
            path: ''
        }
    ]
};

const NoSpec = () => {
    return (
        <NotSelected title="Upload a spec or reconstruct one in order to enable BFLA detection for this API."/>
    );
}

const Learning = () => {
    return (
        <div className="spec-display-wrapper">
            <div className="select-pane">
                <DataCollectionInProgress title="Data collection in progress..." />
            </div>
            <div className="display-pane">
                <div className="in-progress-overlay">
                    <SelectedPathDisplay data={collection_in_progress} />
                </div>
            </div>
        </div>
    );
};

const BflaApiInventory = (props) => {
    const {id: apiId } = props;
    const authModelURL = `modules/bfla/authorizationModel/${apiId}`;
    const [{loading, data}, updateAuthModel] = useFetch(authModelURL);

    if (loading) {
        return <Loader />;
    }
    const {specType, learning} = data || {};

    let specTab;
    if (specType === 'NONE' || !specType) {
        specTab = <NoSpec />;
    }

    if (learning) {
        specTab = <Learning />;
    }

    return (
        <div className="bfla-inventory-wrapper">
            {specTab ||
                <BflaTab data={data} url={authModelURL} loading={loading} refresh={updateAuthModel} />
            }
        </div>
    );
};

const bflaApiInventory = {
    name: 'BFLA',
    component: BflaApiInventory,
    endpoint: '/bfla',
    type: MODULE_TYPES.INVENTORY_DETAILS
};

export default bflaApiInventory;
