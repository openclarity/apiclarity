import React, { useState } from 'react';
import MainTitleWithRefresh from 'components/MainTitleWithRefresh';
import ApiUsage from './ApiUsage';
import MostUsed from './MostUsed';
import LatestSpec from './LatestSpec';

import './dashboard.scss';

const Dashboard = () => {
    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = () => setRefreshTimestamp(Date());

    return (
        <React.Fragment>
            <MainTitleWithRefresh title="Dashboard" onRefreshClick={doRefreshTimestamp} />
            <div className="dashboard-content-wrapper">
                <ApiUsage refreshTimestamp={refreshTimestamp} />
                <div className="dashboard-counters-wrapper">
                    <MostUsed refreshTimestamp={refreshTimestamp} />
                    <LatestSpec refreshTimestamp={refreshTimestamp} />
                </div>
            </div>
        </React.Fragment>
    )
}

export default Dashboard;