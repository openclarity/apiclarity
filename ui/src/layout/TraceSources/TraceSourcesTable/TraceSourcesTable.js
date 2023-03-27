import React, { useMemo } from "react";
import Table from "components/Table";
import { TRACES_TYPES } from "utils/utils";

const TraceSourceTable = ({ filters, refreshTimestamp }) => {
  const columns = useMemo(
    () => [
      {
        Header: "Name",
        id: "name",
        accessor: "name",
      },
      {
        Header: "Type",
        id: "type",
        Cell: ({ row }) => {
          const { type } = row.original;

          if (type) {
            const { label, typeLabel } = TRACES_TYPES[type];

            return (
              <div className="trace-sources-type-cell">
                <span>
                  {typeLabel} | <b>{label}</b>
                </span>
              </div>
            );
          } else {
            return <></>
          }
        },
      },
    ],
    []
  );

  return (
    <Table
      columns={columns}
      paginationItemsName="Traces"
      url="control/traceSources"
      defaultSortBy={[{ id: "name", desc: true }]}
      filters={{
        ...filters,
      }}
      noResultsTitle={`No Trace Sources`}
      refreshTimestamp={refreshTimestamp}
      formatFetchedData={(data) => {
        const { trace_sources } = data || {};

        return { items:trace_sources || [], total: trace_sources?.length || 0 }
      }}
    />
  );
};

export default TraceSourceTable;
