import React from 'react';
import PageContainer from 'components/PageContainer';
import Title from 'components/Title';
import { formatDate } from 'utils/utils';
import ApisList from './ApisList';

const LatestSpec = () => (
    <PageContainer className="latest-spec-wrapper">
        <Title small>Latest spec diffs</Title>
        <ApisList
            url="dashboard/apiUsage/latestDiffs"
            subColumn={{
                title: "Date",
                dataDisplay: ({time}) => <div className="latest-spec-time" style={{whiteSpace: "nowrap"}}>{formatDate(time)}</div>
            }}
            getLink={({apiEventId}) => `/events/tableView/${apiEventId}`}
            apiIdKey="apiEventId"
        />
    </PageContainer>
)

export default LatestSpec;