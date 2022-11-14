import React from 'react';
import { useProgressLoaderReducer, PROGRESS_LOADER_ACTIONS } from 'hooks';
import { FUZZING_STATUS_ITEMS, FUZZING_STATUS_IN_PROGRESS } from './utils';

const FuzzingTestLoaderDisplayWrapper = (props) => {
    const {id, testId, status, startTime, endTime, statusMessage, displayComponent: DisplayComponent,
        inititalLoadStatus=FUZZING_STATUS_ITEMS.IN_PROGRESS.value} = props;

    const isInProgress = FUZZING_STATUS_IN_PROGRESS.includes(status);
    // const baseUrl = `apiSecurity/internalCatalog/${id}`;
    const baseUrl = `modules/fuzzer`;

    const [{loading, customData, status: progressStatus, progress}, dispatch] = useProgressLoaderReducer({
        loadOnMount: isInProgress,
        // statusUrl: `${baseUrl}/fuzzingStatus`,
        statusUrl: `${baseUrl}/report/${id}/${testId}`,
        formatResponse: ({progress, starttime, report={}}) => {
            const {status} = report;
            return {
                status: status || 'IN_PROGRESS',
                progress: progress,
                customData: {statusMessage: '', startTime: starttime, endTime: ''}
            }
        },
        inititalStatus: inititalLoadStatus,
        inProgressStatus: FUZZING_STATUS_ITEMS.IN_PROGRESS.value,
        abortingStatus: FUZZING_STATUS_ITEMS.STOP_IN_PROGRESS.value,
        // needs to be a POST
        abortUrl: `${baseUrl}/fuzz/${id}/stop`
    });

    const {startTime: progressStartTime, endTime: progressEndTime, statusMessage: progressStatusMessage} = customData || {};

    if (loading && !isInProgress) {
        return null;
    }

    const statusProps = !isInProgress ? {status, startTime, endTime, statusMessage} :
        {status: progressStatus, startTime: progressStartTime, endTime: progressEndTime, statusMessage: progressStatusMessage, progress};

    return (
        <DisplayComponent {...statusProps} doAbort={() => dispatch({type: PROGRESS_LOADER_ACTIONS.DO_ABORT})} />
    )
}

export default FuzzingTestLoaderDisplayWrapper;
