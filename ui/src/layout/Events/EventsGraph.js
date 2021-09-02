import React, { useCallback, useEffect } from 'react';
import { useFetch } from 'hooks';
import Chart from 'components/Chart';
import Loader from 'components/Loader';
import COLORS from 'utils/scss_variables.module.scss';

const EventsGraph = ({filters, timeFormat, refreshTimestamp}) => {
    const [{loading, data}, fetchData] = useFetch("apiUsage/hitCount", {loadOnMount: false});
    const doFetchData = useCallback(() => fetchData({queryParams: filters}), [fetchData, filters]);

    useEffect(() => {
        doFetchData();
    }, [doFetchData, refreshTimestamp]);

    return (
        <div className="events-chart-wrapper">
            {loading ? <Loader /> :
                <Chart
                    data={data}
                    configData={[{dataKey: "count", title: "Count", color: COLORS["color-main-light"]}]}
                    timeFormat={timeFormat}
                />
            }
        </div>
    )
}

export default EventsGraph;