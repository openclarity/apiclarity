import COLORS from 'utils/scss_variables.module.scss';
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

export const FUZZING_STATUS_IN_PROGRESS = [FUZZING_STATUS_ITEMS.IN_PROGRESS.value, FUZZING_STATUS_ITEMS.STOP_IN_PROGRESS.value];
