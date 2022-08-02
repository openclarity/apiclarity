import React, { useEffect } from 'react';
import { MODULE_TYPES } from '../MODULE_TYPES.js';

import { useFetch, FETCH_METHODS, usePrevious } from 'hooks';
import Loader from 'components/Loader';
import Button from 'components/Button';
import DownloadJsonButton from 'components/DownloadJsonButton';
import FindingsTable from 'components/FindingsTable';

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
    const annsUrl = `modules/traceanalyzer/apiFindings/${inventoryId}`;
    const [{data}, fetchData] = useFetch(annsUrl);
    const [{loading: deleting}, deleteFinding] = useFetch(annsUrl, {loadOnMount: false});
    const previousDeleting = usePrevious(deleting);

    useEffect(() => {
        if (previousDeleting && !deleting) {
            fetchData();
        }
    }, [previousDeleting, deleting, fetchData]);

    const resetAPIFindings = () => deleteFinding({
        formatUrl: url => `${url}/reset`,
        method: FETCH_METHODS.POST
    });

    return (
        <div className="trace-analysis-wrapper">
            <FindingsTable actionRow={<Button onClick={() => resetAPIFindings()}>Reset Findings</Button>} data={data} />
        </div>
    );
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
