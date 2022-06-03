import React, { useEffect, useMemo, useState } from 'react';
import { useRouteMatch, useHistory } from 'react-router-dom';

import classnames from 'classnames';
import { isNull } from 'lodash';
import Icon, { ICON_NAMES } from 'components/Icon';
import Modal from 'components/Modal';
import Table from 'components/Table';
import Button from 'components/Button';
import LineLoader from 'components/LineLoader';
import DownloadJsonButton from 'components/DownloadJsonButton';
import DropdownSelect from 'components/DropdownSelect';
import { formatDateBy } from 'utils/utils';

import VulnerabilityCounts from './VulnerabilityCounts';

import './tests.scss';
import { useFetch } from 'hooks';


const TestsTable = ({inventoryId, inventoryName}) => {

    const columns = useMemo(() => [
        {
            Header: 'Start time',
            id: "starttime",
            Cell: ({row}) => {
                const {starttime} = row.original;
                return formatDateBy(starttime * 1000, "hh:mm:ss A MMM Do, YYYY");
            },
            canSort: true,
            width: 50
        },
        {
            Header: 'Progress',
            id: "progress",
            Cell: ({row})=> {
                const {progress} = row.original;
                const done = (progress===100);
                return (
                    <div className="fuzzing-line-loader">
                        <LineLoader calculatedPercentage={progress} displayPercent={false} className="fuzzing-line-loader"/>
                        {!done && <div className="fuzzing-line-loader-progress">{progress} %</div>}
                        {done && <Icon
                                name={ICON_NAMES.CHECK_MARK}
                                className={classnames("fuzzing-done-check")}
                            />
                        }
                    </div>
                )
            },
            width: 50
        },
        {
            Header: 'Vulnerabilities',
            id: "vulnerabilities",
            Cell: ({row}) => {
                const {vulnerabilities} = row.original;
                return <VulnerabilityCounts vulnerabilities={vulnerabilities}/>;
            },
        }
    ], []);

    const refreshIntervalInMs = 1000;
    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = () => setRefreshTimestamp(Date());

    const specUrl = "/api/modules/fuzzer/tests/"+inventoryId;
    const [data, setData]=useState({"items":[],"total":0});
    // const [selectedReport, setSelectedReport] = useState();

    const getData = async () => {

        doRefreshTimestamp()
        try {
            fetch(specUrl)
            .then(res => res.json())
            .then(
                (result) => {
                    setData(result);
                },
                (error) => {
                    console.log("fetch error: " + error);
                }
            )
        } catch (err) {
            console.error(err.message);
        }
    };

    useEffect(()=>{

        getData()
        const interval=setInterval(()=>{
            getData()
        },refreshIntervalInMs)
        return()=>clearInterval(interval)
        // eslint-disable-next-line
    },[])

    const [doFuzzAction, setDoFuzzAction] = useState(null);
    const closeResetConfirmationModal = () => setDoFuzzAction(null);

    const [serviceName] = useState(inventoryName.split(".")[0]);
    const [serviceToTest] = useState(serviceName);

    const [{error: fuzzingError}, startFuzzing] = useFetch('modules/fuzzer/fuzz', {loadOnMount: false});

    useEffect(() => {
        if (fuzzingError) {
            console.log('Fuzzing error:', fuzzingError);
        }
    }, [fuzzingError]);

    function DoFuzz(apiID, authDetails={}) {
        startFuzzing({
            formatUrl: url => `${url}/${apiID}`,
            queryParams: {
                ...authDetails
            }
        });
        closeResetConfirmationModal();
    }

    const history = useHistory();
    const {url} = useRouteMatch();

    return (
        <React.Fragment>
            <div className="test-table-wrapper">
                <div className="tests-actions-wrapper">
                    <div className="btn-start-fuzz">
                        <Button onClick={() => setDoFuzzAction(true)}>Start new test </Button>
                    </div>
                    <DownloadJsonButton title="Download findings JSON" fileName="findings-data" data={data} />
                </div>
                <Table
                    columns={columns}
                    paginationItemsName="Tests"
                    //url={`/modules/fuzzer/tests/${inventoryId}`}
                    data={data}
                    defaultSortBy={[{ id: "name", desc: true }]}
                    onLineClick={({ report, starttime }) => history.push(`${url}/${starttime}`)}
                    noResultsTitle={`tests for API '${inventoryName}'`}
                    refreshTimestamp={refreshTimestamp}
                />
            </div>
            {!isNull(doFuzzAction) &&
                <TestingModal
                    title={`Testing API ${inventoryName} (Service '${serviceToTest}')`}
                    onDone={(authDetails) => {
                        DoFuzz(inventoryId, authDetails);
                    }}
                    onClose={closeResetConfirmationModal}
                />
                }
        </React.Fragment>
    )
}

const AUTHENTICATION_METHODS = {
    NONE: {
        value: 'NONE',
        label: 'None',
    },
    BASIC: {
        value: 'BASIC',
        label: 'Basic Auth',
        type: 'basicauth'
    },
    API_KEY: {
        value: 'API_KEY',
        label: 'API Key',
        type: 'apikey'
    },
    BEARER: {
        value: 'BEARER',
        label: 'Bearer Token',
        type: 'bearertoken'
    }
};

const TestingModal = ({title, onClose, onDone}) => {
    const [selectedAuth, setSelectedAuth] = useState(AUTHENTICATION_METHODS.NONE);
    const [authDetails, setAuthDetails] = useState();

    const authItems = Object.values(AUTHENTICATION_METHODS);

    // /fuzz/{apiId}?service={service}&auth={basicauth/apikey/bearertoken}&username={username_encoded}&password={password_encoded}&key={key}&token={token}
    return (
        <Modal
            title={title}
            onClose={onClose}
            className="do-fuzz-confirmation-modal"
            height={600}
            onDone={() => onDone(authDetails)}
            doneTitle="TEST"
        >
            <div>This will test APIs, providing invalid, unexpected or random data as inputs to a computer program.</div>
            <br />
            <div>The APIs are then monitored for exceptions such as crashes, failing built-in code assertions, or potential memory leaks.</div>
            <br />
            <div>Once the testing has completed, the risk findings will be updated accordingly.</div>
            <br />
            <div className="horizontal-box-layout">
                <div><Icon name={ICON_NAMES.ALERT_ROUND} className={classnames("alert-icon")} /></div>
                <b>Testing, in case of anomalies, could crash the API.</b>
            </div>
            <div className="testing-dropdown-title"> <b>Choose an authentication scheme</b></div>
                <DropdownSelect
                    items={authItems}
                    value={selectedAuth}
                    onChange={(selected) => setSelectedAuth(selected)}
                />
            {
                (selectedAuth.value === AUTHENTICATION_METHODS.API_KEY.value &&
                    <ApiKey label={'Key'} onChange={(key) => setAuthDetails(key)} />) ||
                (selectedAuth.value === AUTHENTICATION_METHODS.BEARER.value &&
                    <BearerToken label={'Token'} onChange={(key) => setAuthDetails(key)} />) ||
                (selectedAuth.value === AUTHENTICATION_METHODS.BASIC.value &&
                    <BasicAuth onChange={(auth) => setAuthDetails(auth)} />)
            }
        </Modal>
    );
};

const ApiKey = ({label, onChange}) => {
    const [apiKey, setApiKey] = useState('');
    const [keyValue, setKeyValue] = useState('');
    const {type} = AUTHENTICATION_METHODS.API_KEY;
    useEffect(() => {
        onChange({ type, key: apiKey, value: keyValue});
        // eslint-disable-next-line
    }, [apiKey, keyValue]);
    return (
        <div>
            <div className="input-title">{label}</div>
            <input className="auth-input" value={apiKey} onChange={(e) => setApiKey(e.target.value)} />
            <div className="input-title">Value</div>
            <input className="auth-input" value={keyValue} onChange={(e) => setKeyValue(e.target.value)} />
        </div>
    );
};
const BearerToken = ({label, onChange}) => {
    const [token, setToken] = useState('');
    const {type} = AUTHENTICATION_METHODS.BEARER;
    useEffect(() => {
        onChange({ type, bearertoken: token});
        // eslint-disable-next-line
    }, [token]);
    return (
        <div>
            <div className="input-title">{label}</div>
            <input className="auth-input" value={token} onChange={(e) => setToken(e.target.value)} />
        </div>
    );
};

const BasicAuth = ({onChange}) => {
    const [userName, setUserName] = useState('');
    const [password, setPassword] = useState('');
    const {type} = AUTHENTICATION_METHODS.BASIC;
    useEffect(() => {
        onChange({ type, username: userName, password: password});
        // eslint-disable-next-line
    }, [userName, password]);
    return (
        <div>
            <div className="input-title">User Name</div>
            <input className="auth-input" value={userName} onChange={(e) => setUserName(e.target.value)} />
            <div className="input-title">Password</div>
            <input className="auth-input" value={password} onChange={(e) => setPassword(e.target.value)} />
        </div>
    );
};

export default TestsTable;
