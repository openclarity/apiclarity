import React, { useState } from "react";
import { Route, Switch } from "react-router-dom";
import MainTitleWithRefresh from "components/MainTitleWithRefresh";
import { create } from "context/utils";
import TraceSourcesTable from "./TraceSourcesTable/TraceSourcesTable";
import GeneralFilter, { formatFiltersToQueryParams } from "./GeneralFilter";
import PageContainer from "components/PageContainer";
import Paths from "../../Path";

import "./trace-sources.scss";

const { TRACE_SOURCES_ROOT_PATH } = Paths;

const FILTER_ACTIONS = {
  SET_FILTERS: "SET_FILTERS",
};

const reducer = (state, action) => {
  switch (action.type) {
    case FILTER_ACTIONS.SET_FILTERS: {
      return [...action.payload];
    }
    default:
      return state;
  }
};

const [FilterProvider, useFilterState, useFilterDispatch] = create(reducer, []);

const TraceSources = () => {
  const filters = useFilterState();
  const filterDispatch = useFilterDispatch();
  const setFilters = (filters) =>
    filterDispatch({ type: FILTER_ACTIONS.SET_FILTERS, payload: filters });

  const paramsFilters = formatFiltersToQueryParams(filters);

  const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
  const doRefreshTimestamp = () => setRefreshTimestamp(Date());

  return (
    <div className="inventory-tables-page">
      <MainTitleWithRefresh
        title="Trace Sources"
        onRefreshClick={doRefreshTimestamp}
      />
      <div style={{ display: "flex", gap: "20px" }}>
        <GeneralFilter
          filters={filters}
          onFilterUpdate={(filters) => setFilters(filters)}
        />
      </div>
      <PageContainer className="trace-source-table-page-container">
        <TraceSourcesTable
          filters={paramsFilters}
          refreshTimestamp={refreshTimestamp}
        />
      </PageContainer>
    </div>
  );
};

const TraceSourcesRouter = () => {
  return (
    <FilterProvider>
      <Switch>
        <Route path={TRACE_SOURCES_ROOT_PATH} component={TraceSources} />
      </Switch>
    </FilterProvider>
  );
};

export default TraceSourcesRouter;
