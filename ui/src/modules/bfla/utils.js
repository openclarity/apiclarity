const BFLAState = {
    BFLA_START: 'BFLA_START',
    BFLA_LEARNING: 'BFLA_LEARNING',
    BFLA_DETECTING: 'BFLA_DETECTING',
    BFLA_LEARNT: 'BFLA_LEARNT',
}

const BFLAStatus = {
    NO_SPEC: 'NO_SPEC',
    LEARNING: 'LEARNING',
    LEGITIMATE: 'LEGITIMATE',
    SUSPICIOUS_MEDIUM: 'SUSPICIOUS_MEDIUM',
    SUSPICIOUS_HIGH: 'SUSPICIOUS_HIGH',
}

const BFLA_API_FINDINGS_URL = (apiID) => `/modules/bfla/apiFindings/${apiID}`;
const BFLA_API_AUTH_MODEL_URL = (apiID) => `/modules/bfla/authorizationModel/${apiID}`;
const BFLA_API_AUTH_MODEL_APPROVE_URL = (apiID) => `/modules/bfla/authorizationModel/${apiID}/approve`;
const BFLA_API_AUTH_MODEL_DENY_URL = (apiID) => `/modules/bfla/authorizationModel/${apiID}/deny`;
const BFLA_API_AUTH_MODEL_START_DETECTION_URL = (apiID) => `/modules/bfla/authorizationModel/${apiID}/detection/start`;
const BFLA_API_AUTH_MODEL_START_LEARNING_URL = (apiID) => `/modules/bfla/authorizationModel/${apiID}/learning/start`;
const BFLA_API_AUTH_MODEL_STOP_DETECTION_URL = (apiID) => `/modules/bfla/authorizationModel/${apiID}/detection/stop`;
const BFLA_API_AUTH_MODEL_STOP_LEARNING_URL = (apiID) => `/modules/bfla/authorizationModel/${apiID}/learning/stop`;
const BFLA_API_AUTH_MODEL_RESET_URL = (apiID) => `/modules/bfla/authorizationModel/${apiID}/reset`;
const BFLA_API_AUTH_MODEL_STATE_URL = (apiID) => `/modules/bfla/authorizationModel/${apiID}/state`;
const BFLA_API_API_INVENOTRY_API_INFO = (apiID) => `/apiInventory/${apiID}/apiInfo`;

const BFLA_UTILS = {
    BFLAState,
    BFLAStatus,
    BFLA_API_FINDINGS_URL,
    BFLA_API_AUTH_MODEL_URL,
    BFLA_API_AUTH_MODEL_APPROVE_URL,
    BFLA_API_AUTH_MODEL_DENY_URL,
    BFLA_API_AUTH_MODEL_START_DETECTION_URL,
    BFLA_API_AUTH_MODEL_START_LEARNING_URL,
    BFLA_API_AUTH_MODEL_STOP_DETECTION_URL,
    BFLA_API_AUTH_MODEL_STOP_LEARNING_URL,
    BFLA_API_AUTH_MODEL_RESET_URL,
    BFLA_API_AUTH_MODEL_STATE_URL,
    BFLA_API_API_INVENOTRY_API_INFO
}

export default BFLA_UTILS