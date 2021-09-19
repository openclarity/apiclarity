import React, { useState, useEffect } from 'react';
import classnames from 'classnames';
import { useFetch } from 'hooks';
import TimeFilter, { TIME_SELECT_ITEMS, getTimeFormat } from 'components/TimeFilter';
import Chart from 'components/Chart';
import Loader from 'components/Loader';
import PageContainer from 'components/PageContainer';
import Title from 'components/Title';
import COLORS from 'utils/scss_variables.module.scss';

import './api-usage.scss';

const CONFIG_DATA_MAP = {
    newCount: {dataKey: "newCount", title: "New APIs", color: COLORS["color-main-light"]},
    existingCount: {dataKey: "existingCount", title: "Existing APIs", color: COLORS["color-main"]},
    diffCount: {dataKey: "diffCount", title: "APIs with diffs", color: COLORS["color-status-violet"]}
}

const getCountByTime = (items, time) => {
    const timeItem = items.find(item => item.time === time);

    return !!timeItem ? timeItem.numOfCalls : 0;
}

const UsageChart = ({data, timeFormat}) => {
    const {apisWithDiff=[], existingApis=[], newApis=[]} = data || [];

    const formattedData = [...apisWithDiff].map(apiWithDiff => {
        const {time, numOfCalls} = apiWithDiff;

        return {
            time: time,
            [CONFIG_DATA_MAP.diffCount.dataKey]: numOfCalls,
            [CONFIG_DATA_MAP.existingCount.dataKey]: getCountByTime(existingApis, time),
            [CONFIG_DATA_MAP.newCount.dataKey]: getCountByTime(newApis, time)
        };
    });

    const [activeFilters, setActiveFilters] = useState(Object.keys(CONFIG_DATA_MAP));

    const onFilterClick = (clickedDataKey) => {
        const isActiveFilter = activeFilters.includes(clickedDataKey);
        const newActiveFilters = isActiveFilter ? activeFilters.filter(id => id !== clickedDataKey) : [...activeFilters, clickedDataKey];

        setActiveFilters(newActiveFilters);
    }

    return (
        <div className="upsage-chart-wrapper">
            <div className="chart-filters">
                {
                    Object.values(CONFIG_DATA_MAP).map(({dataKey, title, color}) => (
                        <div
                            key={dataKey}
                            className={classnames("chart-filter", {active: activeFilters.includes(dataKey)})}
                            onClick={() => onFilterClick(dataKey)}
                            style={{color, borderColor: color}}
                        >
                            {title}
                        </div>
                    ))
                }
            </div>
            <Chart
                data={formattedData}
                configData={Object.values(CONFIG_DATA_MAP).filter(configItem => activeFilters.includes(configItem.dataKey))}
                timeFormat={timeFormat}
            />
        </div>
    )
};

const ApiUsage = ({refreshTimestamp}) => {
    const defaultTimeRange = TIME_SELECT_ITEMS.DAY;
    const [timeFilter, setTimeFilter] = useState({selectedRange: defaultTimeRange.value, ...defaultTimeRange.calc()});
    const {selectedRange, startTime, endTime} = timeFilter;

    const [{loading, data}, fetchData] = useFetch("dashboard/apiUsage", {loadOnMount: false});

    useEffect(() => {
        fetchData({queryParams: {startTime, endTime}});
    }, [startTime, endTime, fetchData, refreshTimestamp]);

    return (
        <PageContainer className="api-usage-container">
            <Title>API usage</Title>
            <TimeFilter
                selectedRange={selectedRange}
                startTime={startTime}
                endTime={endTime}
                onChange={({selectedRange, endTime, startTime}) => setTimeFilter({selectedRange, startTime, endTime})}
            />
            {loading ? <Loader /> : <UsageChart data={data} timeFormat={getTimeFormat(startTime, endTime)} />}
        </PageContainer>
    )
}

export default ApiUsage;