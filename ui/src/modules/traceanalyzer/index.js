import React, { useEffect, useState } from 'react';
import { isEmpty } from 'lodash';
import { MODULE_TYPES } from '../MODULE_TYPES.js';

import { useFetch, FETCH_METHODS, usePrevious } from 'hooks';
import Loader from 'components/Loader';
import Table from 'components/Table';
import Button from 'components/Button';
import RiskTag from 'components/RiskTag';
import DownloadJsonButton from 'components/DownloadJsonButton';

import './traceAnalyzer.scss';

const TraceAnalyzerEventDetails = props => {
    const {eventId} = props;
    const [{loading, data}] = useFetch(`modules/traceanalyzer/eventAnnotations/${eventId}`);

    if (loading) {
        return <Loader />;
    }

    if (data && data.items != null) {
        return (
            <div className="findings-details-accordion">
                {data && <TraceDetails trace={data.items} />}
            </div>
        );
    } else {
        return <div className="findings-details-accordion"></div>;
    }
};

const TraceDetails = ({trace}) => (
    <div className="finding-details-wrapper">
        <div className="findings-actions-wrapper">
            <DownloadJsonButton title="Download finding's JSON" fileName="findings-data" data={trace} />
        </div>
        <table className="finding-details-table">
            <thead>
                <tr>
                    <th align="left">Finding</th>
                    <th align="left">Description</th>
                    <th align="left">Risk</th>
                </tr>
            </thead>
            <tbody>
                {
                    trace.map((t, idx) => {
                        return <tr key={idx}>
                            <td>{t.name}</td>
                            {/* <td><TextWithLinks text={description} /></td> */}
                            <td>{t.annotation}</td>
                            <td className={t.severity.toLowerCase()}>{t.severity}</td>
                        </tr>;
                    })
                }
            </tbody>
        </table>
    </div>
);

const TraceAnalyzerAPIDetails = props => {
    const {inventoryId} = props;
    const annsUrl = `modules/traceanalyzer/apiAnnotations/${inventoryId}`;
    const [{data}, fetchData] = useFetch(annsUrl);
    const [selectedRowIds, setSelectedRowIds] = useState([]);

    const columns = [
        { Header: 'Finding',     id: "name",       accessor: "name" },
        { Header: 'Description', id: "annotation", accessor: "annotation" },
        {
            Header: 'Severity',
            Cell: ({row}) => {
                const {severity} = row.original
                return (<RiskTag risk={severity}/>);
            }
        }
    ];

    const [{loading: deleting}, deleteFinding] = useFetch(annsUrl, {loadOnMount: false});
    const previousDeleting = usePrevious(deleting);
    useEffect(() => {
        if (previousDeleting && !deleting) {
            fetchData();
        }
    }, [previousDeleting, deleting, fetchData]);

    const deleteAPIAnnotations = () => deleteFinding({
        //formatUrl: url => `${url}/${resetUrlSuffix}`,
        queryParams: { 'name': data.items[selectedRowIds[0]].kind },
        method: FETCH_METHODS.DELETE
    });

    return (
        <div className="trace-analysis-wrapper">
            <div className="review-actions-wrapper">
                <Button onClick={() => deleteAPIAnnotations()} disabled={isEmpty(selectedRowIds)}>Forget Finding(s)</Button>
                <DownloadJsonButton title="Download finding's JSON" fileName="findings-data" data={data} />
            </div>
            <Table
                noResultsTitle={"this API"}
                columns={columns}
                withPagination={false}
                data={data}
                withMultiSelect={true}
                onRowSelect={setSelectedRowIds}
            />
        </div>
    )
};

const pluginEventDetails = {
    name: 'Trace Analysis',
    moduleName: 'traceanalyzer',
    component: TraceAnalyzerEventDetails,
    endpoint: '/traceanalysis',
    type: MODULE_TYPES.EVENT_DETAILS
};

const pluginAPIDetails = {
    name: 'Trace Analysis',
    moduleName: 'traceanalyzer',
    component: TraceAnalyzerAPIDetails,
    endpoint: '/traceanalysis',
    type: MODULE_TYPES.INVENTORY_DETAILS
}

export {
    pluginEventDetails,
    pluginAPIDetails
}
