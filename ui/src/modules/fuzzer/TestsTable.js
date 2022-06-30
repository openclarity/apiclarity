import React, { useEffect, useMemo, useState } from 'react';
import { useRouteMatch, useHistory } from 'react-router-dom';

import classnames from 'classnames';
import { useFetch, FETCH_METHODS } from 'hooks';
import { isNull } from 'lodash';
import Icon, { ICON_NAMES } from 'components/Icon';
import Modal from 'components/Modal';
import Table from 'components/Table';
import Button from 'components/Button';
import Tooltip from 'components/Tooltip';
import LineLoader from 'components/LineLoader';
import DownloadJsonButton from 'components/DownloadJsonButton';
import DropdownSelect from 'components/DropdownSelect';
import { formatDateBy } from 'utils/utils';

import VulnerabilityCounts from './VulnerabilityCounts';

import COLORS from 'utils/scss_variables.module.scss';

import './tests.scss';

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
                const {progress, errorMessage} = row.original;
                const done = (progress===100);
                const error = (errorMessage && errorMessage.length>0)
                const tooltipId = `spec-diff-${inventoryId}`;
                const tooltipIcon = ICON_NAMES.ALERT;
                const tooltipColor = COLORS["color-error"];
                return (
                    <div className="fuzzing-line-loader">
                        <LineLoader calculatedPercentage={progress} displayPercent={false} className="fuzzing-line-loader" error={error}/>
                        {!done && <div className="fuzzing-line-loader-progress">{progress} %</div>}
                        {done && !error && <Icon
                                name={ICON_NAMES.CHECK_MARK}
                                className={classnames("fuzzing-done-check")}
                            />
                        }
                        {done && error && <Icon
                                name={ICON_NAMES.X_MARK}
                                className={classnames("fuzzing-done-error-check")}
                            />
                        }
                        {error && <div className="spec-diff-icon" style={{width: "22px"}}>
                                <div data-tip data-for={tooltipId}><Icon name={tooltipIcon} style={{tooltipColor}} /></div>
                                <Tooltip id={tooltipId} text={errorMessage} />
                            </div>
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
        // eslint-disable-next-line react-hooks/exhaustive-deps
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

    const [{error: fuzzingError}, startFuzzing] = useFetch(`modules/fuzzer/fuzz/${inventoryId}/start`, {loadOnMount: false});

    useEffect(() => {
        if (fuzzingError) {
            console.log('Fuzzing error:', fuzzingError);
        }
    }, [fuzzingError]);

    function DoFuzz(authDetails, selectedDepth) {
        const depth = selectedDepth ? selectedDepth.value : "QUICK"
        startFuzzing({
            submitData: {
                'auth': authDetails,
                'depth': depth
            },
            headers: { 'Content-Type': 'application/json' },
            method: FETCH_METHODS.POST
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
                    onDone={(authDetails, selectedDepth) => {
                        DoFuzz(authDetails, selectedDepth);
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
        type: 'BasicAuth'
    },
    API_KEY: {
        value: 'API_KEY',
        label: 'API Key',
        type: 'ApiToken'
    },
    BEARER: {
        value: 'BEARER',
        label: 'Bearer Token',
        type: 'BearerToken'
    }
};

const DEPTH_VALUES = {
    QUICK: {
        value: 'QUICK',
        label: 'Quick',
    },
    DEFAULT: {
        value: 'DEFAULT',
        label: 'Default',
    },
    DEEP: {
        value: 'DEEP',
        label: 'Deep',
    }
};

const TestingModal = ({title, onClose, onDone}) => {
    const [selectedAuth, setSelectedAuth] = useState(AUTHENTICATION_METHODS.NONE);
    const [authDetails, setAuthDetails] = useState();

    const [selectedDepth, setSelectedDepth] = useState(DEPTH_VALUES.DEFAULT);

    const authItems = Object.values(AUTHENTICATION_METHODS);
    const depthItems = Object.values(DEPTH_VALUES);

    return (
        <Modal
            title={title}
            onClose={onClose}
            className="do-fuzz-confirmation-modal"
            height={600}
            onDone={() => onDone(authDetails, selectedDepth)}
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
            <div className="testing-dropdown-title"> <b>Choose the test depth</b></div>
                <DropdownSelect
                    items={depthItems}
                    value={selectedDepth}
                    onChange={(selected) => setSelectedDepth(selected)}
                />
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
