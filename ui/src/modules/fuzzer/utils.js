import moment from 'moment';
import { cloneDeep } from 'lodash';
import COLORS from 'utils/scss_variables.module.scss';
import { formatDateBy } from 'utils/utils';
import {asyncDataFetcher} from 'utils/apiUtils';
import { formatDate } from 'utils/utils';

export const AUTH_SCHEME_TYPES = {
    AuthorizationSchemeBasicAuth: {
        value: "BasicAuth",
        label: "Basic Auth"
    },
    AuthorizationSchemeApiToken: {
        value: "ApiToken",
        label: "API Auth"
    },
    AuthorizationSchemeBearerToken: {
        value: "BearerToken",
        label: "Bearer Auth"
    }
};

export const TEST_TYPES = {
    QUICK: {value: "QUICK", label: "Quick scan", duration: "5 minutes."},
    DEFAULT: {value: "DEFAULT", label: "Default scan", duration: "15 minutes"},
    DEEP: {value: "DEEP", label: "Deep scan", duration: "30 minutes"}
};

export const UnfuzzableMessageDisplay = () => (
    <span>Test option disabled in<br />deployments &#62; cluster<br />configuration</span>
)

export const FUZZING_STATUS_ITEMS = {
    IN_PROGRESS: {
        value: "IN_PROGRESS",
        label: "In progress",
        color: COLORS["color-warning-low"],
        getTooltip: ({ progress, startTime }) => (
            <span>
                <b>{`Test in progress ${progress || 0}%`}</b><br />{`started at ${formatDate(startTime)}`}
            </span>
        )
    },
    READY: {
        value: "READY",
        label: "Test-ready",
        color: COLORS["color-main"]
    },
    DONE: {
        value: "DONE",
        label: "Completed",
        color: COLORS["color-success"],
        getTooltip: ({ endTime }) => (
            <span>
                <b>Test completed</b>{` at`}<br />{formatDate(endTime)}
            </span>
        )
    },
    ERROR: {
        value: "ERROR",
        label: "Failed",
        color: COLORS["color-error"],
        getTooltip: ({ statusMessage, endTime }) => (
            <span>
                <b>Test failed</b>{` at ${formatDate(endTime)}`}<br />{statusMessage}
            </span>
        )
    },
    UNFUZZABLE: {
        value: "UNFUZZABLE",
        label: "Not enabled",
        color: COLORS["color-warning"],
        getTooltip: UnfuzzableMessageDisplay
    },
    CANCELLED: {
        value: "CANCELLED",
        label: "Cancelled by user",
        color: COLORS["color-main"]
    },
    STOP_IN_PROGRESS: {
        value: "STOP_IN_PROGRESS",
        label: "Stopping...",
        color: COLORS["color-main"]
    }
};

const getTest = (apiId, starttime) => {
    return new Promise ((resolve) => {
        asyncDataFetcher({url: `modules/fuzzer/report/${apiId}/${starttime}/short`, successCallback: (data) => resolve(data)});
    });
};

const sortTestsDescending = (tests) => {
    return tests.sort((a, b) => {
        return moment.utc(b.starttime).diff(moment.utc(a.starttime));
    });
};

export const convertTestResponse = async (apiId, data) => {
    const testModel = {
        tags: { elements: [], severity: null },
        testDetails: {
            testId: '',
            fuzzingStartTime: null,
            fuzzingStatus: null,
            fuzzingStatusMessage: null,
            testConfiguration: {
                auth: null,
                depth: null
            }
        }
    };

    const getTestStatus = (test) => {
        if (test.errorMessage) {
            return 'ERROR';
        }
        return test.progress === 100 ? 'DONE' : 'IN_PROGRESS';
    };

    let items = data.items || [];
    items = sortTestsDescending(items);
    items = await Promise.all(items.map(async (t) => {
        const report = await getTest(apiId, t.starttime);
        t.tags = report.tags || [];
        t.tags = t.tags.map((t) => {
            t.methods = t.operations.map((op) => {
                return {
                    findings: {
                        elements: op.findings
                    },
                    path: op.operation.path,
                    method: op.operation.method,
                    highestSeverity: op.highestSeverity,
                    requestCount: op.requestsCount
                };
            });

            return t;
        });

        const model = cloneDeep(testModel);
        model.testDetails = {
            ...model.testDetails,
            ...{
                testId: '' + t.starttime,
                fuzzingStartTime: formatDateBy(t.starttime * 1000),
                fuzzingStatus: getTestStatus(t),
            }
        };

        model.tags.highestSeverity = report.highestSeverity;
        model.tags.elements = t.tags;
        return model;
    }));
    return items;
};

export const FUZZING_STATUS_IN_PROGRESS = [FUZZING_STATUS_ITEMS.IN_PROGRESS.value, FUZZING_STATUS_ITEMS.STOP_IN_PROGRESS.value];
