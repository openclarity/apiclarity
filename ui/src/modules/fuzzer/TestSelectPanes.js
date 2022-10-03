import React from 'react';
import DisplaySection from 'components/DisplaySection';
import TitleWithRiskDisplay from 'components/TitleWithRiskDisplay';
import { ICON_NAMES } from 'components/Icon';
import IconWithTitle from 'components/IconWithTitle';
import KeyValueList from 'components/KeyValueList';
import VulnerabilityIcon from 'components/VulnerabilityIcon';
import BreadcrumbSelectPanes, { SelectItemNotification } from 'components/BreadcrumbSelectPanes';
import { TooltipWrapper } from 'components/Tooltip';
// import { FUZZING_STATUS_ITEMS, UnfuzzableMessageDisplay } from 'layout/Apis/utils';
import FindingsAccordion from 'layout/Inventory/InventoryDetails/FindingsAccordion';
import MethodWithRiskDisplay from 'components/MethodWithRiskDisplay';
import { formatDate } from 'utils/utils';
import { SYSTEM_RISKS } from 'utils/systemConsts';
import TestItemDisplay from './TestItemDisplay';
import { TEST_TYPES, AUTH_SCHEME_TYPES, FUZZING_STATUS_ITEMS, UnfuzzableMessageDisplay } from './utils';

const RiskDisplay = ({severity}) => (
    <VulnerabilityIcon severity={ severity || SYSTEM_RISKS.UNKNOWN.value} />
)

const StartTestButton = ({disabled, onClick}) => (
    <IconWithTitle
        name={ICON_NAMES.ADD}
        title="Start new test"
        onClick={onClick}
        disabled={disabled}
    />
)

const TestSelectPanes = ({catalogId, testElements, onNewTestClick, onScanComplete, isFuzzable}) => {
    const selectData = testElements.map(testElement => {
        const {testId, fuzzingStatus, fuzzingStartTime} = testElement.testDetails;

        return {...testElement, id: testId, title: formatDate(fuzzingStartTime), disabled: fuzzingStatus === FUZZING_STATUS_ITEMS.IN_PROGRESS.value};
    });

    console.log(selectData);
    const displayData = [
        {
            getTitle: () => "Tests",
            getSelectItems: () => [...selectData],
            itemDisplay: ({title, testDetails, tags}) => (
                <TestItemDisplay
                    title={title}
                    testDetails={testDetails}
                    catalogId={catalogId}
                    severity={tags.severity || SYSTEM_RISKS.UNKNOWN.value}
                    onScanComplete={onScanComplete}
                />
            ),
            checkAdvanceLevel: () => true,
            customHeaderComponent: () => {
                const testInProgress = testElements[0].testDetails?.fuzzingStatus === FUZZING_STATUS_ITEMS.IN_PROGRESS.value;

                if (testInProgress || !isFuzzable) {
                    return (
                        <TooltipWrapper id="start-test-tooltip" text={testInProgress ? <span>An ongoing test prevents<br />the start of a new test</span> : <UnfuzzableMessageDisplay />}>
                            <StartTestButton disabled />
                        </TooltipWrapper>
                    )
                }

                return (
                    <StartTestButton onClick={onNewTestClick} />
                )
            },
            emptySelectDisplay: () => <SelectItemNotification title="Select a fuzz test to see details" />
        },
        {
            getTitle: () => "Tags",
            getSelectItems: ({tags}) => tags.elements.map(item => ({...item, id: item.name, title: item.name})),
            itemDisplay: ({name, highestSeverity: severity}) => <TitleWithRiskDisplay title={name} risk={severity} customRiskDisplay={() => <RiskDisplay severity={severity} />} />,
            checkAdvanceLevel: () => true,
            emptySelectDisplay: ({title, testDetails}) => {
                const {testConfiguration, fuzzingStatus} = testDetails;
                const {auth, depth} = testConfiguration;
                const {label} = FUZZING_STATUS_ITEMS[fuzzingStatus] || {};

                return (
                    <DisplaySection title={title}>
                        <KeyValueList
                            items={[
                                {label: "Status", value: label},
                                {label: "Test type", value: TEST_TYPES[depth]?.label},
                                {label: "Authorization Scheme", value: AUTH_SCHEME_TYPES[auth?.authorizationSchemeType]?.label}
                            ]}
                        />
                    </DisplaySection>
                )
            }
        },
        {
            getTitle: ({title}) => title,
            getSelectItems: ({methods}) => methods.map(item => ({ ...item, id: `${item.method}-${item.path}` })),
            itemDisplay: ({highestSeverity: severity, method, path}) => (
                <MethodWithRiskDisplay path={path} method={method} risk={severity} customRiskDisplay={() => <RiskDisplay severity={severity} />} />
            ),
            checkAdvanceLevel: () => false,
            emptySelectDisplay: () => <SelectItemNotification title="Select a method to see details" />,
            selectContentDisplay: ({path, method, findings, requestCount}) => (
                <div className="test-findings-content-wrapper">
                    <DisplaySection title={<div className="test-findings-content-title">{`${method} ${path}`}</div>}>
                        <KeyValueList
                            items={[
                                {label: "Number of requests", value: requestCount}
                            ]}
                        />
                    </DisplaySection>
                    <FindingsAccordion findingsDetails={findings} withOccurrencesCount={false} elementsKey="elements" />
                </div>
            )
        },
    ];

    return (
        <BreadcrumbSelectPanes displayData={displayData} mainBreadcrumbsTitle="Tests" />
    )
}

export default TestSelectPanes;
