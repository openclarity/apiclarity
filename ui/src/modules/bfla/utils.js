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

const BFLA_TAB_STATUS = {
    NO_SPEC: 'NO_SPEC',
    DATA_COLLECTION: 'DATA_COLLECTION',
    LEARNING: 'LEARNING',
    DETECTING: 'DETECTING',
    DATA_AVAILABLE: 'DATA_AVAILABLE'
}

function uuidv4() {
    return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, c =>
        // eslint-disable-next-line no-mixed-operators
        (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
    );
}

const formatDataForDisplay = (operations) => {
    const tags = new Set();
    const returnData = {
        tags: []
    };

    // Compute tags
    operations.forEach((operation) => {
        operation.tags.forEach((tag) => tags.add(tag))
    });

    tags.forEach((tag) => returnData.tags.push({ name: tag, authorized: false, paths: [] }))

    // Compute paths
    operations.forEach((operation) => {
        operation.tags.forEach((tag) => {
            const currentTag = returnData.tags.find((returnDataTag) => tag === returnDataTag.name)

            const checkAuthorization = operation.audience.some((audienceMember) => audienceMember.authorized === false)

            currentTag.paths.push({ path: operation.path, method: operation.method, authorized: !checkAuthorization, audience: operation.audience })
        })
    })

    // Check authorization at tag level
    returnData.tags.forEach((tag) => {
        const checkAuthorization = tag.paths.some((path) => path.authorized === false)
        tag.authorized = !checkAuthorization
    });

    return returnData
}

const getDataSelectElements = (data, nestingItem) => data[nestingItem].map((item, key) =>
    ({ ...item, id: key, title: item.name || item.path })
);

const BFLA_API_FINDINGS_URL = (apiID) => `/api/modules/bfla/apiFindings/${apiID}`;
const BFLA_API_AUTH_MODEL_URL = (apiID) => `/api/modules/bfla/authorizationModel/${apiID}`;
const BFLA_API_AUTH_MODEL_APPROVE_URL = (apiID) => `/api/modules/bfla/authorizationModel/${apiID}/approve`;
const BFLA_API_AUTH_MODEL_DENY_URL = (apiID) => `/api/modules/bfla/authorizationModel/${apiID}/deny`;
const BFLA_API_AUTH_MODEL_START_DETECTION_URL = (apiID) => `/api/modules/bfla/authorizationModel/${apiID}/detection/start`;
const BFLA_API_AUTH_MODEL_START_LEARNING_URL = (apiID) => `/api/modules/bfla/authorizationModel/${apiID}/learning/start`;
const BFLA_API_AUTH_MODEL_STOP_DETECTION_URL = (apiID) => `/api/modules/bfla/authorizationModel/${apiID}/detection/stop`;
const BFLA_API_AUTH_MODEL_STOP_LEARNING_URL = (apiID) => `/api/modules/bfla/authorizationModel/${apiID}/learning/stop`;
const BFLA_API_AUTH_MODEL_RESET_URL = (apiID) => `/api/modules/bfla/authorizationModel/${apiID}/reset`;
const BFLA_API_AUTH_MODEL_STATE_URL = (apiID) => `/api/modules/bfla/authorizationModel/${apiID}/state`;
const BFLA_API_API_INVENOTRY_API_INFO = (apiID) => `/apiInventory/${apiID}/apiInfo`;

const BFLA_UTILS = {
    BFLAState,
    BFLAStatus,
    BFLA_TAB_STATUS,
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
    BFLA_API_API_INVENOTRY_API_INFO,
    uuidv4,
    formatDataForDisplay,
    getDataSelectElements
}

export default BFLA_UTILS
