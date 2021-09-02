import React from 'react';
import Title from 'components/Title';
import ApiUsage from './ApiUsage';
import MostUsed from './MostUsed';
import LatestSpec from './LatestSpec';

import './dashboard.scss';

const Dashboard = () => {
    return (
        <React.Fragment>
            <Title>Dashboard</Title>
            <div className="dashboard-content-wrapper">
                <ApiUsage />
                <div className="dashboard-counters-wrapper">
                    <MostUsed />
                    <LatestSpec />
                </div>
            </div>
        </React.Fragment>
    )
}

export default Dashboard;