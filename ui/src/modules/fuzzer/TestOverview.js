import React, { useEffect, useState } from 'react';
import { useParams, useRouteMatch, useHistory } from 'react-router-dom';
import {useFetch} from 'hooks';
import Table from 'components/Table';
import Loader from 'components/Loader';
import BackRouteButton from 'components/BackRouteButton';
import DownloadJsonButton from 'components/DownloadJsonButton';
import { formatDateBy  } from 'utils/utils';

import TestDetails from './TestDetails';
import VulnerabilityCounts from './VulnerabilityCounts';

const TestOverview = ({backUrl}) => {
    const {startTime, inventoryId} = useParams();
    const specUrl = `modules/fuzzer/report/${inventoryId}/${startTime}`;
    const [report, setReport] = useState();
    const [{loading, data}] = useFetch(specUrl);
    const [testDetails, setTestDetails] = useState();

    const {url} = useRouteMatch();
    const history = useHistory();

    useEffect(() => {
        if (data) {
            const {report: topLevelReport} = data || {};
            const {report} = topLevelReport || {report:[]};
            const formattedReport = Object.keys(report).map((k) => report[k]);
            setReport({ items: formattedReport });
        }
    },[data, startTime]);

    const columns = [
        { Header: 'Name', id: "name", accessor: "name" },
        { Header: 'Test Type', id: "testType", accessor: "testType" },
        // { Header: 'Test Type', id: "testType", Cell: ({row}) => console.log(row) },
        { Header: 'Description', id: "description", accessor: "description" },
        {
            Header: 'Requests Performed',
            id: "requests",
            Cell: ({row}) => {
                const {paths} = row.original;
                return paths.length;
            },
            width: 50
        },
        {
            Header: 'Findings',
            id: "findings",
            Cell: ({ row }) => {
                const { findings } = row.original;
                const findingResults = findings.map((f) => f.request)
                      .reduce((accum, f) => {
                          if (accum[f.severity] || accum[f.severity] === 0) {
                              accum[f.severity]++;
                              accum.total++;
                          }
                          return accum;
                      }, { total:0, critical: 0, high: 0, medium: 0, low: 0 });

                return <VulnerabilityCounts vulnerabilities={findingResults} />;
            }
        },
    ];

    if (loading) {
        return <Loader />;
    }

    if (testDetails) {
        return <TestDetails test={testDetails} goBack={() => history.push(url)} />;
    }

    const TestTitle = `Test Report ${formatDateBy(startTime * 1000, "hh:mm:ss A MMM Do, YYYY")}:`;
    return (
        <div className="test-table-wrapper">
            <div className="tests-actions-wrapper">
                <BackRouteButton title={TestTitle} path={backUrl} />
                <DownloadJsonButton title="Download findings JSON" fileName="findings-data" data={report} />
            </div>
            <Table
                noResultsTitle={"this event"}
                columns={columns}
                data={report}
                onLineClick={(test) => setTestDetails(test)}
            />
        </div>
    );
};

export default TestOverview;
