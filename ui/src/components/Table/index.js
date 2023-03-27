import React, { useMemo, useCallback, useEffect, useState } from "react";
import classnames from "classnames";
import { isEmpty, isEqual, pickBy, isNull, isUndefined } from "lodash";
import {
  useTable,
  usePagination,
  useSortBy,
  useResizeColumns,
  useFlexLayout,
  useRowSelect,
  useExpanded,
} from "react-table";
import Icon, { ICON_NAMES } from "components/Icon";
import IconWithTitle from "components/IconWithTitle";
import Loader from "components/Loader";
import Checkbox from "components/Checkbox";
import NoResultsDisplay from "components/NoResultsDisplay";
import Arrow from "components/Arrow";
import { useFetch, usePrevious } from "hooks";
import Pagination from "./Pagination";
import ColumnsSelectPanel from "./ColumnsSelectPanel";
import * as utils from "./utils";

import "./table.scss";

export { utils };

const SELECTOR_COLUMN_ID = "SELECTOR";
const EXPANDER_COLUMN_ID = "EXPANDER";

const SELECTOR_WIDTH = 40;
const EXPANDER_WIDTH = 40;

const RowSelectCheckbox = React.memo(
  ({ checked, indeterminate, onChange }) => (
    <Checkbox
      checked={checked || indeterminate}
      onChange={onChange}
      halfSelected={indeterminate}
    />
  ),
  (prevProps, nextProps) => {
    const { checked: prevChecked, indeterminate: prevIndeterminate } =
      prevProps;
    const { checked, indeterminate } = nextProps;

    const shouldRefresh =
      checked !== prevChecked || indeterminate !== prevIndeterminate;

    return !shouldRefresh;
  }
);

const ColumnPanelSelect = ({ toggleColumns }) => {
  return (
    <IconWithTitle
      className="table-columns-class"
      name={ICON_NAMES.COLUMNS}
      title="Columns"
      onClick={toggleColumns}
    />
  );
};

const Table = (props) => {
  const {
    columns,
    defaultSortBy: defaultSortByItems,
    onLineClick,
    paginationItemsName,
    url,
    formatFetchedData,
    filters,
    hideColumnControl = true,
    noResultsTitle = "API",
    refreshTimestamp,
    withPagination = true,
    data: externalData,
    withMultiSelect = false,
    onRowSelect,
    markedRowIds = [],
    innerRowComponent: InnerRowComponent,
  } = props;

  const [{ loading, data, error }, fetchData] = useFetch(url, {
    loadOnMount: false,
    formatFetchedData,
  });
  const fetchedData = !!formatFetchedData ? formatFetchedData(data) : data;
  const dataJson = JSON.stringify(fetchedData || []);
  const externalDataJson = JSON.stringify(externalData);
  const tableData = useMemo(
    () => (!!url ? JSON.parse(dataJson) : JSON.parse(externalDataJson)),
    [url, dataJson, externalDataJson]
  );
  const defaultSortBy = useMemo(
    () => defaultSortByItems || [],
    [defaultSortByItems]
  );

  const defaultColumn = React.useMemo(
    () => ({
      minWidth: 30, // minWidth is only used as a limit for resizing,
      width: 100,
    }),
    []
  );

  const {
    getTableProps,
    getTableBodyProps,
    headerGroups,
    prepareRow,
    page,
    allColumns,
    canPreviousPage,
    nextPage,
    previousPage,
    gotoPage,
    toggleSortBy,
    state: { pageIndex, pageSize, sortBy, selectedRowIds },
  } = useTable(
    {
      columns,
      getRowId: (rowData, rowIndex) => (!!rowData.id ? rowData.id : rowIndex),
      data: tableData?.items || [],
      defaultColumn,
      initialState: {
        pageIndex: 0,
        pageSize: 50,
        sortBy: defaultSortBy,
        selectedRowIds: {},
      },
      manualPagination: true,
      pageCount: -1,
      manualSortBy: true,
      disableMultiSort: true,
    },
    useResizeColumns,
    useFlexLayout,
    useSortBy,
    useExpanded,
    usePagination,
    useRowSelect,
    (hooks) => {
      hooks.useInstanceBeforeDimensions.push(({ headerGroups }) => {
        // fix the parent group of the framework columns to not be resizable
        headerGroups[0].headers.forEach((header) => {
          if (header.originalId === SELECTOR_COLUMN_ID) {
            header.canResize = false;
          }
        });
      });

      hooks.visibleColumns.push((columns) => {
        const updatedColumns = columns;

        if (withMultiSelect) {
          updatedColumns.unshift({
            Header: ({ getToggleAllRowsSelectedProps }) => (
              <RowSelectCheckbox {...getToggleAllRowsSelectedProps()} />
            ),
            id: SELECTOR_COLUMN_ID,
            Cell: ({ row }) => (
              <div className="table-row-checkbox-wrapper">
                <RowSelectCheckbox {...row.getToggleRowSelectedProps()} />
              </div>
            ),
            disableResizing: true,
            minWidth: SELECTOR_WIDTH,
            width: SELECTOR_WIDTH,
            maxWidth: SELECTOR_WIDTH,
          });
        }
        if (!!InnerRowComponent) {
          updatedColumns.unshift({
            Header: () => null, // No header
            id: EXPANDER_COLUMN_ID,
            Cell: ({ row }) => {
              const { onClick } = row.getToggleRowExpandedProps();

              return (
                <Arrow
                  name={row.isExpanded ? "top" : "bottom"}
                  className="row-expand-icon"
                  onClick={onClick}
                  small
                />
              );
            },
            disableResizing: true,
            minWidth: EXPANDER_WIDTH,
            width: EXPANDER_WIDTH,
            maxWidth: EXPANDER_WIDTH,
          });
        }

        return updatedColumns;
      });
    }
  );

  const { id: sortKey, desc: sortDesc } = !isEmpty(sortBy) ? sortBy[0] : {};
  const cleanFilters = pickBy(
    filters,
    (value) => !isNull(value) && value !== ""
  );
  const prevCleanFilters = usePrevious(cleanFilters);
  const filtersChanged = !isEqual(cleanFilters, prevCleanFilters);
  const prevPageIndex = usePrevious(pageIndex);
  const prevSortKey = usePrevious(sortKey);
  const prevSortDesc = usePrevious(sortDesc);
  const sortingChanged = sortKey !== prevSortKey || sortDesc !== prevSortDesc;

  const getQueryParams = useCallback(() => {
    const queryParams = {
      ...cleanFilters,
    };

    if (withPagination) {
      queryParams.page = pageIndex + 1;
      queryParams.pageSize = pageSize;
    }

    if (!isEmpty(sortKey)) {
      queryParams.sortKey = sortKey;
      queryParams.sortDir = sortDesc ? "DESC" : "ASC";
    }

    return queryParams;
  }, [pageIndex, pageSize, sortKey, sortDesc, cleanFilters, withPagination]);

  const doFetchWithQueryParams = useCallback(() => {
    if (loading) {
      return;
    }

    fetchData({ queryParams: getQueryParams() });
  }, [fetchData, getQueryParams, loading]);

  useEffect(() => {
    if (!filtersChanged && pageIndex === prevPageIndex && !sortingChanged) {
      return;
    }

    if (filtersChanged && pageIndex !== 0) {
      gotoPage(0);

      return;
    }

    if (!!url) {
      doFetchWithQueryParams();
    }
  }, [
    filtersChanged,
    pageIndex,
    prevPageIndex,
    doFetchWithQueryParams,
    gotoPage,
    sortingChanged,
    refreshTimestamp,
    url,
  ]);

  const selectedRows = Object.keys(selectedRowIds);
  const prevSelectedRows = usePrevious(selectedRows);

  useEffect(() => {
    if (!!onRowSelect && !isEqual(selectedRows, prevSelectedRows)) {
      onRowSelect(selectedRows);
    }
  }, [prevSelectedRows, selectedRows, onRowSelect]);

  const [showColumnsPanel, setShowColumnsPanel] = useState(false);

  if (!!error) {
    return null;
  }
  const headerColumnNames = (
    headerGroups.length > 1
      ? headerGroups[0].headers.map((header) => {
          // if (STATIC_HEADER_IDS.includes(header.originalId)) {
          //     return null;
          // }

          return header.Header;
        })
      : []
  ).filter((item) => !isNull(item));

  const resultsTotalText = isUndefined(tableData.total)
    ? ""
    : `Showing ${tableData.total} entries`;

  return (
    <div className="table-wrapper">
      {!hideColumnControl && (
        <ColumnPanelSelect
          toggleColumns={() => setShowColumnsPanel(!showColumnsPanel)}
        />
      )}
      {showColumnsPanel && (
        <ColumnsSelectPanel
          headerColumnNames={headerColumnNames}
          columns={allColumns}
          onClose={() => setShowColumnsPanel(false)}
          columnsIconClassName={"table-columns-class"}
        />
      )}
      {!withPagination ? (
        <div className="no-pagination-results-total">{resultsTotalText}</div>
      ) : (
        <Pagination
          canPreviousPage={canPreviousPage}
          nextPage={nextPage}
          previousPage={previousPage}
          pageIndex={pageIndex}
          pageSize={pageSize}
          displayName={paginationItemsName}
          gotoPage={gotoPage}
          loading={loading}
          total={tableData?.total || 0}
          page={page}
        />
      )}
      <div className="table" {...getTableProps()}>
        <div className="table-head">
          {headerGroups.map((headerGroup) => {
            return (
              <div className="table-tr" {...headerGroup.getHeaderGroupProps()}>
                {headerGroup.headers.map((column) => {
                  const { id, isSorted, isSortedDesc } = column;

                  return (
                    <div className="table-th" {...column.getHeaderProps()}>
                      {column.render("Header")}
                      {column.canSort && (
                        <Icon
                          className={classnames(
                            "table-sort-icon",
                            { sorted: isSorted },
                            { rotate: isSortedDesc && isSorted }
                          )}
                          name={ICON_NAMES.SORT}
                          onClick={() =>
                            toggleSortBy(id, isSorted ? !isSortedDesc : false)
                          }
                        />
                      )}
                      {column.canResize && (
                        <div
                          {...column.getResizerProps()}
                          className={classnames("resizer", {
                            isResizing: column.isResizing,
                          })}
                        />
                      )}
                    </div>
                  );
                })}
              </div>
            );
          })}
        </div>
        <div className="table-body" {...getTableBodyProps()}>
          {isEmpty(page) && !loading && (
            <NoResultsDisplay
              title={`No results available for ${noResultsTitle}`}
            />
          )}
          {loading && (
            <div className="table-loading">
              <Loader />
            </div>
          )}
          {page.map((row) => {
            prepareRow(row);

            return (
              <React.Fragment key={row.id}>
                <div
                  className={classnames(
                    "table-tr",
                    { clickable: !!onLineClick },
                    { marked: markedRowIds.includes(row.id) },
                    { expanded: row.isExpanded }
                  )}
                  {...row.getRowProps()}
                >
                  {row.cells.map((cell) => {
                    const { className } = cell.column;
                    const cellClassName = classnames("table-td", {
                      [className]: className,
                    });

                    const isTextValue = !!cell.column.accessor;

                    return (
                      <div
                        className={cellClassName}
                        {...cell.getCellProps()}
                        onClick={() => {
                          if (!!onLineClick) {
                            onLineClick(row.original);
                          }
                        }}
                      >
                        {isTextValue ? cell.value : cell.render("Cell")}
                      </div>
                    );
                  })}
                </div>
                {row.isExpanded && !!InnerRowComponent && (
                  <div className="table-tr inner-row">
                    <div className="table-td">
                      <InnerRowComponent data={row.original} />
                    </div>
                  </div>
                )}
              </React.Fragment>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default React.memo(Table, (prevProps, nextProps) => {
  const {
    filters: prevFilters,
    refreshTimestamp: prevRefreshTimestamp,
    data: prevData,
    markedRowIds: prevMarkedRowIds,
  } = prevProps;
  const { filters, refreshTimestamp, data, markedRowIds } = nextProps;

  const shouldRefresh =
    !isEqual(prevFilters, filters) ||
    prevRefreshTimestamp !== refreshTimestamp ||
    !isEqual(prevData, data) ||
    !isEqual(prevMarkedRowIds, markedRowIds);

  return !shouldRefresh;
});
