import React, { useState, useEffect, useCallback } from 'react';
import { useFetch } from 'hooks';
import TimeFilter, { TIME_SELECT_ITEMS, getTimeFormat } from 'components/TimeFilter';
import Chart from 'components/Chart';
import Loader from 'components/Loader';
import { SPEC_TYPES } from '../utils';

import COLORS from 'utils/scss_variables.module.scss';

import './method-hit-count.scss';

const SPEC_TYPE_TO_PATH_FILTER= {
    [SPEC_TYPES.PROVIDED]: "providedPathID[is]",
    [SPEC_TYPES.RECONSTRUCTED]: "reconstructedPathID[is]"
};

const MethodHitCount = ({method, pathId, specType, spec}) => {
    const defaultTimeRange = TIME_SELECT_ITEMS.DAY;
    const [timeFilter, setTimeFilter] = useState({selectedRange: defaultTimeRange.value, ...defaultTimeRange.calc()});
    const {selectedRange, startTime, endTime} = timeFilter;

    const [{loading, data}, fetchData] = useFetch("apiUsage/hitCount", {loadOnMount: false});
    const doFetchHitCount = useCallback(({queryParams}) =>
        fetchData({queryParams: {...queryParams, "method[is]": [method], [SPEC_TYPE_TO_PATH_FILTER[specType]]: pathId, "spec[is]": spec, "showNonApi": false}}),
        [fetchData, method, pathId, specType, spec]);

    useEffect(() => {
        doFetchHitCount({queryParams: {startTime, endTime}});
    }, [startTime, endTime, doFetchHitCount]);

    return (
        <div className="method-hit-count-wrapper">
            <div className="hit-count-title">Hit count</div>
            <div className="hit-count-content">
                <TimeFilter
                    selectedRange={selectedRange}
                    startTime={startTime}
                    endTime={endTime}
                    onChange={({selectedRange, endTime, startTime}) => setTimeFilter({selectedRange, startTime, endTime})}
                />
                <div className="hit-count-chart-wrapper">
                    {loading ? <Loader /> :
                        <Chart
                            data={data}
                            configData={[{dataKey: "count", title: "Count", color: COLORS["color-main"]}]}
                            timeFormat={getTimeFormat(startTime, endTime)}
                        />
                    }
                </div>
            </div>
        </div>
    )
}

export default MethodHitCount;