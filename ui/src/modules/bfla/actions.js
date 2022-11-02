import axios from 'axios';
import utils from './utils';

const getBflaApiFindings = (apiID) => axios.get(utils.BFLA_API_FINDINGS_URL(apiID));
const getBflaApiAuthModel = (apiID) => axios.get(utils.BFLA_API_AUTH_MODEL_URL(apiID));
const postBflaApiAuthModel = (apiID, body) => axios.post(utils.BFLA_API_AUTH_MODEL_URL(apiID), body);
const putBflaApiAuthModelApprove = (apiID, method, path, k8sClientUid) => axios.put(utils.BFLA_API_AUTH_MODEL_APPROVE_URL(apiID), null, { params: { method, path, k8sClientUid } });
const putBflaApiAuthModelDeny = (apiID, method, path, k8sClientUid) => axios.put(utils.BFLA_API_AUTH_MODEL_DENY_URL(apiID), null, { params: { method, path, k8sClientUid } });
const putBflaApiAuthModelStartDetection = (apiID) => axios.put(utils.BFLA_API_AUTH_MODEL_START_DETECTION_URL(apiID));
const putBflaApiAuthModelStartLearning = (apiID) => axios.put(utils.BFLA_API_AUTH_MODEL_START_LEARNING_URL(apiID));
const putBflaApiAuthModelStopDetection = (apiID) => axios.put(utils.BFLA_API_AUTH_MODEL_STOP_DETECTION_URL(apiID));
const putBflaApiAuthModelStopLearning = (apiID) => axios.put(utils.BFLA_API_AUTH_MODEL_STOP_LEARNING_URL(apiID));
const postBflaApiAuthModelReset = (apiID) => axios.post(utils.BFLA_API_AUTH_MODEL_RESET_URL(apiID));
const getBflaApiAuthModelState = (apiID) => axios.get(utils.BFLA_API_AUTH_MODEL_STATE_URL(apiID));
const getBflaApiInfoForNoSpec = (apiID) => axios.get(utils.BFLA_API_API_INVENOTRY_API_INFO(apiID))

const BFLA_ACTIONS = {
    getBflaApiFindings,
    getBflaApiAuthModel,
    postBflaApiAuthModel,
    putBflaApiAuthModelApprove,
    putBflaApiAuthModelDeny,
    putBflaApiAuthModelStartDetection,
    putBflaApiAuthModelStartLearning,
    putBflaApiAuthModelStopDetection,
    putBflaApiAuthModelStopLearning,
    postBflaApiAuthModelReset,
    getBflaApiAuthModelState,
    getBflaApiInfoForNoSpec,
}

export default BFLA_ACTIONS
