import React, {useState} from 'react';
import { isEmpty } from 'lodash';
import ListDisplay from 'components/ListDisplay';
import Arrow, { ARROW_NAMES } from 'components/Arrow';
import DownloadJsonButton from 'components/DownloadJsonButton';

import emptySelectImage from 'utils/images/select.svg';

const NotSelected = ({title}) => (
    <div className="not-selected-wrapper">
        <div className="not-selected-title">{title}</div>
        <img src={emptySelectImage} alt="no tag selected" />
    </div>
);

const BackHeader = ({title, onBack}) => (
    <div className="selected-back-header">
        <Arrow name={ARROW_NAMES.LEFT} onClick={onBack} />
        <div>{title}</div>
    </div>
);

const Label = ({title}) => {
    return (
        <div className="header-label">
            {title}
        </div>
    )
};

const TestDetailsDisplay = ({data, testType, description}) => {
    const {type, paths, findings} = data || {};
    return (
        <div className="finding-details-wrapper">
            <div className="findings-actions-wrapper">
                <div className="label-wrapper">
                    <Label title={`Test Type: ${testType}`} />
                    <Label title={`Description: ${description}`} />
                </div>
                <DownloadJsonButton title="Download findings JSON" fileName="findings-data" data={data} />
            </div>
            {type === 'PATHS' &&
                <table className="finding-details-table">
                    <thead>
                        <tr>
                            <th align="left">CALL</th>
                            <th align="left">RETURN CODE</th>
                            <th align="left">DOWNLOAD JSON</th>
                        </tr>
                    </thead>
                    <tbody>
                        {
                            paths.map((p, idx) => {
                                return <tr key={idx}>
                                    <td>{p.verb} {p.uri}</td>
                                    <td>{p.result}</td>
                                    <td><DownloadJsonButton title="" fileName="findings-data" data={p} /></td>
                                </tr>;
                            })
                        }
                    </tbody>
                </table>
            }
            {type === 'FINDINGS' &&
                <table className="finding-details-table">
                    <thead>
                        <tr>
                            <th align="left">NAME</th>
                            <th align="left">DESCRIPTION</th>
                            <th align="left">RISK</th>
                        </tr>
                    </thead>
                    <tbody>
                        {
                            findings.map((f, idx) => {
                                return <tr key={idx}>
                                    <td>{f.type}</td>
                                    <td>{f.description}</td>
                                    <td className={`${f.request.severity}`}>{f.request.severity}</td>
                                </tr>;
                            })
                        }
                    </tbody>
                </table>
            }
        </div>
    );
};

const TestDetails = ({test, goBack}) => {
    const [selectedItem, setSelectedItem] = useState();
    const {name: testName} = test;
    const list = [{id: 'paths', type: 'PATHS', name: 'HTTP Calls', paths: test.paths}, {id: 'findings', type: 'FINDINGS', name: 'Findings List', findings: test.findings}];
    return (
        <div className="spec-display-wrapper">
            <div className="select-pane">
                <BackHeader title={testName} onBack={goBack} />
                <ListDisplay
                    items={list}
                    itemDisplay={({name}) => <div>{name}</div>}
                    onSelect={(selected) => setSelectedItem(selected)}
                    selectedId={selectedItem ? selectedItem.id : null}
                />
            </div>
            <div className="display-pane">
                {isEmpty(selectedItem) ? <NotSelected title={<span>Select a report to see details.</span>} /> :
                    <TestDetailsDisplay data={selectedItem} testType={test.testType} description={test.description} />}
            </div>
        </div>
    );
};

export default TestDetails;
