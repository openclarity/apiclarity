import React from 'react';
import PageContainer from 'components/PageContainer';
import Title from 'components/Title';
import Tag from 'components/Tag';
import ApisList from './ApisList';

const MostUsed = ({refreshTimestamp}) => (
    <PageContainer className="most-used-wrapper">
        <Title small>Most used APIs</Title>
        <ApisList
            url="dashboard/apiUsage/mostUsed"
            subColumn={{
                title: "Calls (No.)",
                dataDisplay: ({numCalls}) => <Tag rounded>{numCalls}</Tag>
            }}
            getLink={({apiType, apiInfoId}) => `/inventory/${apiType}/${apiInfoId}`}
            apiIdKey="apiInfoId"
            refreshTimestamp={refreshTimestamp}
        />
    </PageContainer>
)

export default MostUsed;