import React, { useEffect } from 'react';
import { usePrevious } from 'hooks';
import VulnerabilityIcon from 'components/VulnerabilityIcon';
import Icon, { ICON_NAMES } from 'components/Icon';
import { TooltipWrapper } from 'components/Tooltip';
import Button from 'components/Button';
import DotLoader from 'components/DotLoader';
import { FUZZING_STATUS_ITEMS, FUZZING_STATUS_IN_PROGRESS } from '../utils';
import FuzzingTestLoaderDisplayWrapper from '../FuzzingTestLoaderDisplayWrapper';

import COLORS from 'utils/scss_variables.module.scss';

import './test-item-display.scss';

const InProgressDisplay = ({id, progress, doAbort, allowStop}) => (
    <div style={{display: "flex", alignItems: "center"}}>
        {allowStop && <div style={{marginRight: "10px"}} onClick={() => console.log('click')}><Button secondary onClick={doAbort}>Stop test</Button></div>}
        <TooltipWrapper id={id} text={<span><b>Test in progress</b>{` ${progress}%`}</span>}><DotLoader /></TooltipWrapper>
    </div>
)

const DoneDisplay = ({severity}) => (
    <VulnerabilityIcon severity={severity} />
)

const ErrorDisplay = ({id, statusMessage, isCancelled}) => (
    <TooltipWrapper id={id} text={isCancelled ? "Test stopped by user" : <span style={{whiteSpace: "break-spaces"}}><b>Test failed</b>{` ${statusMessage}`}</span>}>
        <Icon name={ICON_NAMES.X_MARK} style={{color: COLORS["color-error"]}} />
    </TooltipWrapper>
)

const checkInProgress = status => FUZZING_STATUS_IN_PROGRESS.includes(status);

const TestDisplay = ({id, title, severity, status, statusMessage, progress, onScanComplete, doAbort}) => {
    const prevStatus = usePrevious(status);

    const isInProgress = checkInProgress(status);
    const isDone = status === FUZZING_STATUS_ITEMS.DONE.value;
    const StatusDisplay = (isInProgress || !status) ? InProgressDisplay : (isDone ? DoneDisplay : ErrorDisplay);

    useEffect(() => {
        if (checkInProgress(prevStatus) && !isInProgress) {
            onScanComplete();
        }
    }, [prevStatus, status, isInProgress, onScanComplete]);

    return (
        <div className="fuzzing-test-display-item">
            <div>{title}</div>
            <StatusDisplay
                id={id}
                progress={progress}
                severity={severity}
                isCancelled={status === FUZZING_STATUS_ITEMS.CANCELLED.value}
                statusMessage={statusMessage}
                doAbort={doAbort}
                allowStop={status === FUZZING_STATUS_ITEMS.IN_PROGRESS.value}
            />
        </div>
    )
}

const TestItemDisplay = ({catalogId, title, testDetails, severity, onScanComplete}) => {
    const {testId, fuzzingStatus, fuzzingStartTime, fuzzingStatusMessage} = testDetails;

    return (
        <FuzzingTestLoaderDisplayWrapper
            id={catalogId}
            status={fuzzingStatus}
        // currently the id is the timestamp,
            testId={testId}
            startTime={fuzzingStartTime}
            statusMessage={fuzzingStatusMessage}
            displayComponent={props => (
                <TestDisplay {...props} id={testId} title={title} severity={severity} onScanComplete={onScanComplete} />
            )}
            inititalLoadStatus={null}
        />
    );
}

export default TestItemDisplay;
