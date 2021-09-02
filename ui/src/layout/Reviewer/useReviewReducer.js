import { useReducer, useEffect } from 'react';
import { usePrevious, useFetch } from 'hooks';

export const initialState = {
    isLoading: true,
    isLoadingError: false,
    dataToReview: [],
    reviewId: null,
    openParamInputPathData: null,
    mergingPathsData: null
};

export const REVIEW_ACTIONS = {
    DATA_LOADED: "DATA_LOADED",
    ERROR_LOADIND_DATA: "ERROR_LOADIND_DATA",
    UPDATE_REVIEW_DATA: "UPDATE_REVIEW_DATA",
    SET_OPEN_PARAM_INPUT_PATH: "SET_OPEN_PARAM_INPUT_PATH",
    SET_MERGING_PATHS_DATA: "SET_MERGING_PATHS_DATA"
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
                reviewId
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
        default:
            return state;
    }
}

function useReviewReducer({inventoryId}) {
    const [reviewState, dispatch] = useReducer(reducer, initialState);
    const {isLoading, isLoadingError, dataToReview, reviewId, openParamInputPathData, mergingPathsData} = reviewState;

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


    return [{loading: isLoading, isLoadingError, dataToReview, reviewId, openParamInputPathData, mergingPathsData}, dispatch];
}

export default useReviewReducer;