import React, { useMemo } from 'react';
import { useHistory } from 'react-router-dom';
import Table, { utils } from 'components/Table';

const InventoryTable = ({basePath, type, filters, refreshTimestamp}) => {
    const columns = useMemo(() => [
        {
            Header: 'Name',
            id: "name",
            accessor: "name"
        },
        {
            Header: 'Port',
            id: "port",
            accessor: "port",
            width: 30
        },
        {
            Header: 'Trace Source',
            id: "traceSource",
            Cell: ({row}) => {
                const {traceSourceName, traceSourceType} = row.original;

                return (
                    <div>
                        {traceSourceName} - {traceSourceType}
                    </div>
                )
            },
            width: 30
        },
        {
            Header: 'Provided Spec',
            id: "hasProvidedSpec",
            Cell: ({row}) => {
                const {hasProvidedSpec} = row.original;

                return (
                    <utils.StatusIndicatorIcon isSuccess={hasProvidedSpec} />
                )
            },
            canSort: true,
            width: 30
        },
        {
            Header: 'Reconstructed Spec',
            id: "hasReconstructedSpec",
            Cell: ({row}) => {
                const {hasReconstructedSpec} = row.original;

                return (
                    <utils.StatusIndicatorIcon isSuccess={hasReconstructedSpec} />
                )
            },
            canSort: true,
            width: 30
        }
    ], []);

    const history = useHistory();

    return (
        <Table
            columns={columns}
            paginationItemsName="APIs"
            url="apiInventory"
            defaultSortBy={[{id: "name", desc: true}]}
            filters={{
                type,
                ...filters
            }}
            onLineClick={({id}) => history.push(`${basePath}/${type}/${id}`)}
            noResultsTitle={`${type.toLowerCase()} APIs`}
            refreshTimestamp={refreshTimestamp}
            formatFetchedData={(data) => {
                const { items } = data || {};

                return { items: items || [],  total: items?.length || 0 }
            }}
        />
    )
}

export default InventoryTable;