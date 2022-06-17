import React, { useMemo } from 'react';
import Table from 'components/Table';
import RiskTag from 'components/RiskTag';
import DownloadJsonButton from 'components/DownloadJsonButton';

import './tests.scss';

const FindingsTable = ({inventoryId, inventoryName}) => {
    const columns = useMemo(() => [
        {
            Header: 'Name',
            id: "name",
            accessor: "name",
            canSort: true,
            width: 50
        },
        {
            Header: 'Description',
            id: "description",
            accessor: "description",
        },
        {
            Header: 'Risk',
            id: "risk",
            Cell: ({row}) => {
                const {risk} = row.original;
                return (<div className="risk-tag-color"><RiskTag risk={risk.toUpperCase()}/></div>);
            },
            canSort: true,
            width: 20
        }
    ], []);

    return (
        <div className="findings-table-wrapper">
            <div className="findings-actions-wrapper">
                <DownloadJsonButton title="Download findings JSON" fileName="findings-data" data={{}} />
            </div>
            <Table
                columns={columns}
                paginationItemsName="Findings"
                url={`/modules/fuzzer/findings/${inventoryId}`}
                defaultSortBy={[{ id: "name", desc: true }]}
                noResultsTitle={`findings for API '${inventoryName}'`}
            />
        </div>
    )
}

export default FindingsTable;
