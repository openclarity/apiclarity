import React, { useEffect, useState, useCallback } from 'react';
import { Route, Switch, useRouteMatch, Redirect, useLocation } from 'react-router-dom';
import MainTitleWithRefresh from 'components/MainTitleWithRefresh';
import Icon, { ICON_NAMES } from 'components/Icon';
import TabbedPageContainer from 'components/TabbedPageContainer';
import TimeFilter, { TIME_SELECT_ITEMS, getTimeFormat } from 'components/TimeFilter';
import ToggleButton from 'components/ToggleButton';
import { usePrevious } from 'hooks';
import { create } from 'context/utils';
import EventsTable from './EventsTable';
import EventsGraph from './EventsGraph';
import EventDetails from './EventDetails';
import GeneralFilter, { formatFiltersToQueryParams } from './GeneralFilter';

import './events.scss';

const FILTER_ACTIONS = {
    SET_GENERAL: "SET_GENERAL",
    SET_TIME: "SET_TIME",
    SET_SHOW_NON_API: "SET_SHOW_NON_API"
};

const reducer = (state, action) => {
    switch (action.type) {
        case FILTER_ACTIONS.SET_GENERAL: {
            return {
                ...state,
                generalFilters: action.payload
            };
        }
        case FILTER_ACTIONS.SET_TIME: {
            return {
                ...state,
                timeFilter: action.payload
            };
        }
        case FILTER_ACTIONS.SET_SHOW_NON_API: {
            return {
                ...state,
                showNonApi: action.payload
            };
        }
        default:
            return state;
    }
}

const defaultTimeRange = TIME_SELECT_ITEMS.DAY;
const [FilterProvider, useFilterState, useFilterDispatch] = create(reducer, {
    generalFilters: [],
    timeFilter: {selectedRange: defaultTimeRange.value, ...defaultTimeRange.calc()},
    showNonApi: false
});

const getTableViewPath = path => `${path}/tableView`;
const getGraphViewPath = path => `${path}/graphView`;

const TabTitle = ({icon, title}) => (
    <div className="events-tab-title">
        <Icon name={icon} />
        <span>{title}</span>
    </div>
);

const Events = () => {
    const {path} = useRouteMatch();

    const {pathname} = useLocation();
    const prevPathname = usePrevious(pathname);
    
    const {generalFilters: filters, timeFilter, showNonApi} = useFilterState();
    
    const filterDispatch = useFilterDispatch();
    const setTimeFilter = useCallback((timeFilter) => filterDispatch({type: FILTER_ACTIONS.SET_TIME, payload: timeFilter}), [filterDispatch]);
    const setFilters = (filters) => filterDispatch({type: FILTER_ACTIONS.SET_GENERAL, payload: filters});
    const setShowNonApi = (showNonApi) => filterDispatch({type: FILTER_ACTIONS.SET_SHOW_NON_API, payload: showNonApi});

    const {selectedRange, startTime, endTime} = timeFilter;
    const refreshTimeFilter = useCallback(() => {
        const selectedRangeItem = TIME_SELECT_ITEMS[selectedRange];

        if (!!selectedRangeItem.calc) {
            setTimeFilter({selectedRange, ...selectedRangeItem.calc()});
        }
    }, [setTimeFilter, selectedRange]);

    const paramsFilters = formatFiltersToQueryParams(filters);

    const reuqestFilters = {startTime, endTime, ...paramsFilters, showNonApi};

    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = () => setRefreshTimestamp(Date());

    const doRefresh = () => {
        const selectedRangeItem = TIME_SELECT_ITEMS[selectedRange];

        if (!selectedRangeItem.calc) {
            doRefreshTimestamp();

            return;
        }

        refreshTimeFilter();
    }

    useEffect(() => {
        if (prevPathname !== pathname) {
            refreshTimeFilter();
        }
    }, [prevPathname, pathname, refreshTimeFilter]);

    return (
        <div className="events-page">
            <MainTitleWithRefresh title="API Events" onRefreshClick={doRefresh} />
            <div className="events-filters-wrapper">
                <TimeFilter
                    selectedRange={selectedRange}
                    startTime={startTime}
                    endTime={endTime}
                    onChange={({selectedRange, endTime, startTime}) => setTimeFilter({selectedRange, startTime, endTime})}
                />
                <GeneralFilter
                    filters={filters}
                    onFilterUpdate={filters => setFilters(filters)}
                    customFilterDisplay={() => <ToggleButton title="Show non-API events" value={showNonApi} onChange={setShowNonApi} />}
                />
            </div>
            <TabbedPageContainer
                items={[
                    {
                        title: <TabTitle title="Table view" icon={ICON_NAMES.TABLE} />,
                        to: getTableViewPath(path),
                        component: () => <EventsTable filters={reuqestFilters} refreshTimestamp={refreshTimestamp} />
                    },
                    {
                        title: <TabTitle title="Graph view" icon={ICON_NAMES.GRAPH} />,
                        to: getGraphViewPath(path),
                        component: () => <EventsGraph filters={reuqestFilters} timeFormat={getTimeFormat(startTime, endTime)} refreshTimestamp={refreshTimestamp} />
                    }
                ]}
            />
        </div>
    )
};

const EventsRouter = () => {
    const {path} = useRouteMatch();
    const tablePath = getTableViewPath(path);

    return (
        <FilterProvider>
            <Switch>
                <Redirect exact from="/events" to={tablePath} />
                <Route path={`${tablePath}/:eventId`} component={EventDetails} />
                <Route path={path} component={Events} />
            </Switch>
        </FilterProvider>
    )
}

export default EventsRouter;