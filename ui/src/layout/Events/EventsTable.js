import React, { useMemo } from 'react';
import { useHistory, useRouteMatch } from 'react-router-dom';
import Table, { utils } from 'components/Table';
import MethodTag from 'components/MethodTag';
import StatusIndicator from 'components/StatusIndicator';
import RiskTag from 'components/RiskTag';
import SpecDiffIcon, { SPEC_DIFF_TYPES_MAP } from 'components/SpecDiffIcon';
import { formatDate } from 'utils/utils';
import { API_TYPE_ITEMS } from 'layout/Inventory';

const EventsTable = ({filters, refreshTimestamp}) => {
    const columns = useMemo(() => [
        {
            Header: 'Time',
            id: "time",
            accessor: original => formatDate(original.time),
            width: 70,
            alwaysShow: true
        },
        {
            Header: 'Method',
            id: "method",
            Cell: ({row}) => {
                const {method} = row.original;

                return (
                    <MethodTag method={method} />
                )
            },
            canSort: true,
            width: 40,
            alwaysShow: true
        },
        {
            Header: 'Path',
            id: "path",
            accessor: "path",

            alwaysShow: true
        },
        {
            Header: 'Status Code',
            id: "statusCode",
            Cell: ({row}) => {
                const {statusCode} = row.original;

                return (
                    <StatusIndicator title={statusCode} isError={statusCode >= 400} />
                )
            },
            canSort: true,
            width: 40,
            alwaysShow: true
        },
        {
            Header: 'Source',
            id: "sourceIP",
            accessor: "sourceIP",
            width: 50
        },
        {
            Header: 'Destination',
            id: "destinationIP",
            accessor: "destinationIP",
            width: 50
        },
        {
            Header: 'Destination Port',
            id: "destinationPort",
            accessor: "destinationPort",
            width: 55
        },
        {
            Header: 'Spec Diff',
            id: "specDiffType",
            Cell: ({row}) => {
                const {id, specDiffType} = row.original;

                const {value} = SPEC_DIFF_TYPES_MAP[specDiffType] || {};

                if (!value || value === SPEC_DIFF_TYPES_MAP.NO_DIFF.value) {
                    return <utils.EmptyValue />;
                }

                return (
                    <SpecDiffIcon id={id} specDiffType={specDiffType} />
                )
            },
            canSort: true,
            width: 40
        },
        {
            Header: 'Host',
            id: "hostSpecName",
            accessor: "hostSpecName",
            width: 30
        },
        {
            Header: 'Type',
            id: "apiType",
            accessor: original => {
                const typeItem = API_TYPE_ITEMS[original.apiType];

                return !!typeItem ? typeItem.label : null;
            },
            width: 30
        },
        {
            Header: 'Alerts',
            id: "alerts",
            accessor: original => {
                if (original.alerts == null) {
                    return "";
                }
                const alerts = original.alerts.map((a, idx) => {
                    const alert = a.reason.split('_')[1];
                    return <RiskTag key={idx} risk={alert} label={a.moduleName} />;
                });
                return <div className="alerts-column">{alerts}</div>;
            },
            width: 50
        }
    ], []);

    const history = useHistory();
    const {path} = useRouteMatch();

    return (
        <Table
            hideColumnControl={false}
            columns={columns}
            paginationItemsName="APIs"
            url="apiEvents"
            defaultSortBy={[{id: "time", desc: true}]}
            filters={filters}
            onLineClick={({id}) => history.push(`${path}/${id}`)}
            noResultsTitle="API events"
            refreshTimestamp={refreshTimestamp}
            formatFetchedData={(data) => {
                const { items } = data || {};

                return { items: items || [],  total: items?.length || 0 }
            }}
        />
    )
}

export default EventsTable;
