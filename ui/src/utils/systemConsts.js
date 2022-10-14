export const RULE_ACTIONS = {
    DETECT: "DETECT",
    BLOCK: "BLOCK",
    ALLOW: "ALLOW",
    ENCRYPT: "ENCRYPT",
    ENCRYPT_DIRECT: "ENCRYPT_DIRECT"
};

export const RISKY_ACTION_ITEM = {value: "RISKY", label: "Risky"};

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
}

export const VULNERABILITY_LEVELS = {
    [SYSTEM_RISKS.CRITICAL.value]: SYSTEM_RISKS.CRITICAL,
    [SYSTEM_RISKS.HIGH.value]: SYSTEM_RISKS.HIGH,
    [SYSTEM_RISKS.MEDIUM.value]: SYSTEM_RISKS.MEDIUM,
    [SYSTEM_RISKS.LOW.value]: SYSTEM_RISKS.LOW,
    [SYSTEM_RISKS.UNKNOWN.value]: SYSTEM_RISKS.UNKNOWN
};

export const API_RISK_ITEMS = {
    [SYSTEM_RISKS.CRITICAL.value]: SYSTEM_RISKS.CRITICAL,
    [SYSTEM_RISKS.HIGH.value]: SYSTEM_RISKS.HIGH,
    [SYSTEM_RISKS.MEDIUM.value]: SYSTEM_RISKS.MEDIUM,
    [SYSTEM_RISKS.LOW.value]: SYSTEM_RISKS.LOW,
    [SYSTEM_RISKS.NEUTRAL.value]: SYSTEM_RISKS.NEUTRAL,
    [SYSTEM_RISKS.UNKNOWN.value]: SYSTEM_RISKS.UNKNOWN,
    [SYSTEM_RISKS.NO_RISK.value]: SYSTEM_RISKS.NO_RISK,
}
