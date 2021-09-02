import React from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { formatDate, formatDateBy } from 'utils/utils';

import COLORS from 'utils/scss_variables.module.scss';
import './chart.scss';

const CHART_MARGIN = 20;
const AXIS_TICK_STYLE = {fill: COLORS["color-grey"], fontSize: "14px"};

const Chart = ({data, configData, timeFormat="HH:mm:ss"}) => (
    <ResponsiveContainer width="100%" height="100%">
        <AreaChart
            data={data}
            margin={{top: CHART_MARGIN, right: CHART_MARGIN, left: CHART_MARGIN, bottom: CHART_MARGIN}}
        >
            <defs>
                {
                    configData.map(({dataKey, color}) => (
                        <linearGradient key={dataKey} id={dataKey} x1="0" y1="0" x2="0" y2="1">
                            <stop offset="0.6%" stopColor={color} stopOpacity={0.2}/>
                            <stop offset="49%" stopColor={color} stopOpacity={0}/>
                        </linearGradient>
                    ))
                }
            </defs>
            <CartesianGrid stroke={COLORS["color-grey-lighter"]} vertical={false} />
            <XAxis dataKey="time" stroke={COLORS["color-grey-light"]} tick={AXIS_TICK_STYLE} tickFormatter={time => formatDateBy(time, timeFormat)} />
            <YAxis stroke="transparent" tick={AXIS_TICK_STYLE} />
            <Tooltip content={({payload, active}) => {
                if (!active || !payload) {
                    return null;
                }
                
                return (
                    <div className="chart-tooltip-content">
                        {
                            configData.map(({dataKey, title}) => {
                                const {value} = payload.find(payloadItem => payloadItem.dataKey === dataKey);
                                
                                return (
                                    <div key={dataKey}>{`${title}: ${value}`}</div>
                                )
                            })
                        }
                        <div>{`Time: ${formatDate(payload[0].payload.time)}`}</div>
                    </div>
                )
            }} />
            {
                configData.map(({dataKey, color}) => (
                    <Area key={dataKey} type="linear" dataKey={dataKey} stroke={color} fillOpacity={1} fill={`url(#${dataKey})`} />
                ))
            }
        </AreaChart>
    </ResponsiveContainer>
);

export default Chart;