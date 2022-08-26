import moment from 'moment';

export const formatDateBy = (date, format) => !!date ? moment(date).format(format): "";
export const formatDate = (date) => formatDateBy(date, "MMM Do, YYYY HH:mm:ss");

export const SYSTEM_RISKS = {
    CRITICAL: {value: "CRITICAL", label: "Critical"},
    HIGH: {value: "HIGH", label: "High"},
    MEDIUM: {value: "MEDIUM", label: "Medium"},
    LOW: {value: "LOW", label: "Low"},
    UNKNOWN: {value: "UNKNOWN", label: "Unknown"},
    NO_RISK: {value: "NO_RISK", label: "No known risk"},
    APPROVED: {value: "APPROVED", label: "Approved"},
    NEUTRAL: {value: "NEUTRAL", label: "Neutral"},
    INFO: {value: "INFO", label: "Info"}
};
