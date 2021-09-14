import React from 'react';
import PageContainer from 'components/PageContainer';
import Title from 'components/Title';
import SpecDiffIcon from 'components/SpecDiffIcon';
import { formatDateBy } from 'utils/utils';
import ApisList from './ApisList';

const LatestSpec = ({refreshTimestamp}) => (
    <PageContainer className="latest-spec-wrapper">
        <Title small>Latest spec diffs</Title>
        <ApisList
            url="dashboard/apiUsage/latestDiffs"
            getLink={({apiEventId}) => `/events/tableView/${apiEventId}`}
            apiIdKey="apiEventId"
            refreshTimestamp={refreshTimestamp}
            columnItems={[
                {title: "Type", content: ({apiEventId, diffType}) => <SpecDiffIcon id={apiEventId} specDiffType={diffType} />},
                {title: "Date", content: ({time}) => <div className="latest-spec-time" style={{whiteSpace: "nowrap"}}>{formatDateBy(time, "MMM Do, HH:mm")}</div>}
            ]}
        />
    </PageContainer>
)

export default LatestSpec;