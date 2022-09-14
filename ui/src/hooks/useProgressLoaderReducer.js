import { useReducer, useEffect, useRef } from 'react';
import { usePrevious, useFetch, FETCH_METHODS } from 'hooks';

const initialState = {
    isLoading: false,
    isLoadingError: false,
    loadData: false,
    progress: null,
    customData: null,
    doAbort: false,
    doLoadStatus: false
};

export const LOADER_ACTIONS = {
    ERROR_LOADIND_STATUS: "ERROR_LOADIND_STATUS",
    STATUS_DATA_LOADED: "STATUS_DATA_LOADED",
    DO_LOAD_STATUS: "DO_LOAD_STATUS",
    DO_ABORT: "DO_ABORT"
}

const getReducer = ({inProgressStatus, abortingStatus}) => (
    (state, action) => {
        switch (action.type) {
            case LOADER_ACTIONS.STATUS_DATA_LOADED: {
                const {status, progress, customData} = action.payload;

                return {
                    ...state,
                    isLoading: [inProgressStatus, abortingStatus].includes(status),
                    status,
                    progress,
                    customData,
                    doAbort: false,
                    doLoadStatus: false
                };
            }
            case LOADER_ACTIONS.ERROR_LOADIND_STATUS: {
                return {
                    ...state,
                    isLoading: false,
                    isLoadingError: true
                };
            }
            case LOADER_ACTIONS.DO_ABORT: {
                return {
                    ...state,
                    isLoading: true,
                    status: abortingStatus,
                    doAbort: true
                };
            }
            case LOADER_ACTIONS.DO_LOAD_STATUS: {
                return {
                    ...state,
                    isLoading: true,
                    status: inProgressStatus,
                    doLoadStatus: true
                };
            }
            default:
                return state;
        }
    }
)

//IMPORTANT: formatResponse needs to return a {status, progress, customData} object
function useProgressLoaderReducer({loadOnMount=true, statusUrl, abortUrl, formatResponse, inititalStatus="NONE", inProgressStatus="IN_PROGRESS", abortingStatus="ABORTING"}) {
    const [specState, dispatch] = useReducer(getReducer({inProgressStatus, abortingStatus}), {...initialState, status: inititalStatus});
    const {isLoading, isLoadingError, loadData, customData, status, progress, doAbort, doLoadStatus} = specState;
    const prevDoAbort = usePrevious(doAbort);
    const prevDoLoadStatus = usePrevious(doLoadStatus);

    const [{loading, data, error}, fetchStatus] = useFetch(statusUrl, {loadOnMount});
    const prevLoading = usePrevious(loading);

    const [{loading: aborting, error: abortError}, abortReconstruct] = useFetch(abortUrl, {loadOnMount: false});
    const prevAborting = usePrevious(aborting);

    const fetcherRef = useRef(null);

    useEffect(() => {
        return function cleanup() {
            if (fetcherRef.current) {
                clearTimeout(fetcherRef.current);
            }
        };
    }, []);

    useEffect(() => {
        if (prevLoading && !loading) {
            if (!!error) {
                dispatch({type: LOADER_ACTIONS.ERROR_LOADIND_STATUS});
            } else {
                const formattedResponseData = formatResponse(data);

                dispatch({type: LOADER_ACTIONS.STATUS_DATA_LOADED, payload: formattedResponseData});

                if ([inProgressStatus, abortingStatus].includes(formattedResponseData.status)) {
                    fetcherRef.current = setTimeout(() => fetchStatus(), 3000);
                }
            }
        }
    }, [prevLoading, loading, data, error, fetchStatus, abortingStatus, formatResponse, inProgressStatus]);

    useEffect(() => {
        if (!prevDoAbort && doAbort) {
            clearTimeout(fetcherRef.current);
            abortReconstruct({method: FETCH_METHODS.POST});
        }
    }, [prevDoAbort, doAbort, abortReconstruct]);

    useEffect(() => {
        if (!prevDoLoadStatus && doLoadStatus) {
            fetchStatus();
        }
    }, [prevDoLoadStatus, doLoadStatus, fetchStatus]);

    useEffect(() => {
        if (prevAborting && !aborting && !abortError) {
            fetchStatus()
        }
    }, [prevAborting, aborting, abortError, fetchStatus]);

    useEffect(() => {
        if (!loadData) {
            return;
        }

        fetchStatus();
    }, [fetchStatus, loadData]);

    return [{loading: isLoading, isLoadingError, customData, status, progress}, dispatch];
}

export default useProgressLoaderReducer;
