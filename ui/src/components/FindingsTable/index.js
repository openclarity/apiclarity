import React, {useMemo} from 'react';
import Table from 'components/Table';
import DownloadJsonButton from 'components/DownloadJsonButton';
import VulnerabilityIcon from 'components/VulnerabilityIcon';
import RoundIconContainer from 'components/RoundIconContainer';
import {ICON_NAMES} from 'components/Icon';
import {SYSTEM_RISKS} from 'utils/utils';

import './findings-table.scss';

const FindingsRisk = ({ severity }) => {
    const isInfo = severity === SYSTEM_RISKS.INFO.value;
    const {label} = SYSTEM_RISKS[severity] || {};
    const riskClass = severity.toLowerCase();

    return (
        <div className="vulnerability-item">
            {isInfo ? <RoundIconContainer name={ICON_NAMES.INFO} /> :
                <VulnerabilityIcon severity={severity} />
            } <div className={`risk-text ${riskClass}`}>{label}</div>
        </div>
    );
};

const FindingsInnerTable = ({data}) => {
    const {additional_info} = data;
    const additionInfo = JSON.stringify(additional_info, null, 4);

    return (
        <div className="findings-inner-table">
            <div className="findings-inner-table-title">
                Additional Info
            </div>
            <pre>
                {additionInfo ? additionInfo : '{}'}
            </pre>
        </div>
    );
};

const FindingsTable = ({columns: userColumns=[], data={}, actionRow, url=''}) => {
    const columns = useMemo(() => [{
            Header: 'Name',
            id: "name",
            accessor: 'name',
            width: 50,
        },
        {
            Header: 'Description',
            id: 'description',
            accessor: "description",
        },
        ...userColumns,
        {
            Header: 'Risk',
            id: 'severity',
            width: 40,
            Cell: ({row}) => {
                const { severity } = row.original;
                return <FindingsRisk severity={severity} />;
            }
        }], [userColumns]);

    return (
        <div className="findings-table-wrapper">
            <div className="findings-actions-wrapper">
                {actionRow}
                <DownloadJsonButton title="Download findings JSON" fileName="findings-data" data={data} />
            </div>
            <Table
                columns={columns}
                paginationItemsName="Findings"
                url={url}
                data={{ items: data.items}}
                defaultSortBy={[{ id: "name", desc: true }]}
                innerRowComponent={FindingsInnerTable}
                withPagination={false}
                formatFetchedData={(data) => {
                    const { items } = data || {};

                    return { items: items || [],  total: items?.length || 0 }
                }}
                /* noResultsTitle={`findings for API '${inventoryName}'`} */
            />
        </div>
    );
};

export default FindingsTable;
