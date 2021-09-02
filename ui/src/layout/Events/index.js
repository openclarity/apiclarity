import React, { useState } from 'react';
import { Route, Switch, useRouteMatch, Redirect } from 'react-router-dom';
import Title from 'components/Title';
import Icon, { ICON_NAMES } from 'components/Icon';
import TabbedPageContainer from 'components/TabbedPageContainer';
import TimeFilter, { TIME_SELECT_ITEMS, getTimeFormat } from 'components/TimeFilter';
import ToggleButton from 'components/ToggleButton';
import EventsTable from './EventsTable';
import EventsGraph from './EventsGraph';
import EventDetails from './EventDetails';
import GeneralFilter, { formatFiltersToQueryParams } from './GeneralFilter';

import './events.scss';

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

    const defaultTimeRange = TIME_SELECT_ITEMS.DAY;
    const [timeFilter, setTimeFilter] = useState({selectedRange: defaultTimeRange.value, ...defaultTimeRange.calc()});
    const {selectedRange, startTime, endTime} = timeFilter;

    const [filters, setFilters] = useState([]);
    const paramsFilters = formatFiltersToQueryParams(filters);

    const [showNonApi, setShowNonApi] = useState(false);

    const reuqestFilters = {startTime, endTime, ...paramsFilters, "showNonApi": showNonApi};

    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = () => setRefreshTimestamp(Date());

    const doRefresh = () => {
        const selectedRangeItem = TIME_SELECT_ITEMS[selectedRange];

        if (!selectedRangeItem.calc) {
            doRefreshTimestamp();

            return;
        }

        setTimeFilter({selectedRange, ...selectedRangeItem.calc()})
    }

    return (
        <div className="events-page">
            <div className="events-page-title">
                <Title>API Events</Title>
                <Icon name={ICON_NAMES.REFRESH} onClick={doRefresh} />
            </div>
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
        <Switch>
            <Redirect exact from="/events" to={tablePath} />
            <Route path={`${tablePath}/:eventId`} component={EventDetails} />
            <Route path={path} component={Events} />
        </Switch>
    )
}

export default EventsRouter;