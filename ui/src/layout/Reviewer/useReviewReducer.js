import { useReducer, useEffect } from 'react';
import { usePrevious, useFetch } from 'hooks';
import { FILTERS_MAP } from './GeneralFilter';

export const initialState = {
    isLoading: true,
    isLoadingError: false,
    dataToReview: [],
    reviewId: null,
    openParamInputPathData: null,
    mergingPathsData: null,
    filters: []
};

export const REVIEW_ACTIONS = {
    DATA_LOADED: "DATA_LOADED",
    ERROR_LOADIND_DATA: "ERROR_LOADIND_DATA",
    UPDATE_REVIEW_DATA: "UPDATE_REVIEW_DATA",
    SET_OPEN_PARAM_INPUT_PATH: "SET_OPEN_PARAM_INPUT_PATH",
    SET_MERGING_PATHS_DATA: "SET_MERGING_PATHS_DATA",
    UPDATE_PATH_PARAM_NAME: "UPDATE_PATH_PARAM_NAME",
    SET_FILTERS: "SET_FILTERS"
}

const filterData = (data, filters=[]) => {
    let filteredData = [...(data || [])];

    filters.forEach(({scope, value}) => {
        filteredData = filteredData.filter(({suggestedPath, apiEventsPaths}) => {
            if (scope === FILTERS_MAP.path.value) {
                return !!value.find(path => suggestedPath.includes(path));
            } else if (scope === FILTERS_MAP.method.value) {
                const pathMethods = [...new Set(apiEventsPaths.reduce((acc, curr) => ([...acc, ...curr.methods]), []))];
                
                return !!value.find(method => pathMethods.includes(method))
            }

            return false;
        });
    });

    return filteredData;
}

const getFormatDataWithIds = data => data.map((item, index) => ({...item, id: String(index)}));

const reducer = (state, action) => {
    switch (action.type) {
        case REVIEW_ACTIONS.DATA_LOADED: {
            const {id: reviewId, reviewPathItems} = action.payload;

            return {
                ...state,
                isLoading: false,
                dataToReview: getFormatDataWithIds(reviewPathItems) || [],
                reviewId,
                filters: []
            };
        }
        case REVIEW_ACTIONS.ERROR_LOADIND_DATA: {
            return {
                isLoading: false,
                isLoadingError: true
            }
        }
        case REVIEW_ACTIONS.UPDATE_REVIEW_DATA: {
            return {
                ...state,
                dataToReview: getFormatDataWithIds(action.payload),
                mergingPathsData: null,
                openParamInputPathData: null
            };
        }
        case REVIEW_ACTIONS.SET_OPEN_PARAM_INPUT_PATH: {
            return {
                ...state,
                openParamInputPathData: action.payload
            };
        }
        case REVIEW_ACTIONS.SET_MERGING_PATHS_DATA: {
            return {
                ...state,
                mergingPathsData: action.payload
            };
        }
        case REVIEW_ACTIONS.UPDATE_PATH_PARAM_NAME: {
            const {dataToReview} = state;
            const {updatedPath, pathData} = action.payload;

            const updatingPathIndex = dataToReview.findIndex(({suggestedPath}) => suggestedPath === pathData.suggestedPath);
            
            return {
                ...state,
                dataToReview: [
                    ...dataToReview.slice(0, updatingPathIndex),
                    {...pathData, suggestedPath: updatedPath},
                    ...dataToReview.slice(updatingPathIndex + 1)
                ],
                openParamInputPathData: null
            }
        }
        case REVIEW_ACTIONS.SET_FILTERS: {
            return {
                ...state,
                filters: action.payload
            };
        }
        default:
            return state;
    }
}

function useReviewReducer({inventoryId}) {
    const [reviewState, dispatch] = useReducer(reducer, initialState);
    const {isLoading, isLoadingError, dataToReview, reviewId, openParamInputPathData, mergingPathsData, filters} = reviewState;

    const [{loading, data, error}] = useFetch(`apiInventory/${inventoryId}/suggestedReview`, {loadOnMount: !!inventoryId});
    const prevLoading = usePrevious(loading);
    
    useEffect(() => {
        if (prevLoading && !loading) {
            if (!!error) {
                dispatch({type: REVIEW_ACTIONS.ERROR_LOADIND_DATA});
            } else {
                dispatch({type: REVIEW_ACTIONS.DATA_LOADED, payload: data});
            }
        }
    }, [prevLoading, loading, data, error]);


    return [{
        loading: isLoading,
        isLoadingError,
        dataToReview,
        reviewId,
        openParamInputPathData,
        mergingPathsData,
        filters,
        filteredData: filterData(dataToReview, filters)
    }, dispatch];
}

export default useReviewReducer;