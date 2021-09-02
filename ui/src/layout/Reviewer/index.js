import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useLocation, Redirect, useHistory } from 'react-router-dom';
import { isEmpty, isNull } from 'lodash';
import { useFetch, FETCH_METHODS, usePrevious } from 'hooks';
import { useNotificationDispatch, showNotification } from 'context/NotificationProvider'; 
import BackRouteButton from 'components/BackRouteButton';
import Title from 'components/Title';
import Table from 'components/Table';
import PageContainer from 'components/PageContainer';
import Button from 'components/Button';
import Tag from 'components/Tag';
import Loader from 'components/Loader';
import { SPEC_TAB_ITEMS } from 'layout/Inventory/InventoryDetails/Specs'
import PathDisplay from './PathDisplay';
import MergeModal from './MergeModal';
import ConfirmationModal from './ConfirmationModal';
import useReviewReducer, { REVIEW_ACTIONS } from './useReviewReducer';
import { getPathWithParamInIndex, getMethodsFromPaths } from './utils';

import './reviewer.scss';

const Bold = ({children}) => <span style={{fontWeight: "bold"}}>{children}</span>

const Reviewer = () => {
    const {query} = useLocation();
    const {inventoryId, inventoryName, backUrl} = query || {};

    const history = useHistory();
    const backQuery = useMemo(() => ({inititalSelectedTab: SPEC_TAB_ITEMS.RECONSTRUCTED.value}), []);
    const returnToInventory = useCallback(() => history.push({pathname: backUrl, query: backQuery}), [history, backUrl, backQuery]);

    const [{loading, isLoadingError, reviewId, dataToReview, openParamInputPathData, mergingPathsData}, dispatch] = useReviewReducer({inventoryId});

    const [selectedRowIds, setSelectedRowIds] = useState([]);
    const selectedRowsCount = selectedRowIds.length;

    const [showConfirmationModal, setShowConfirmationModal] = useState(false);
    const closeConfirmationModal = () => setShowConfirmationModal(false);

    const notificationDispatch = useNotificationDispatch();
    const showReviewCompletedNotification = useCallback(() => showNotification(notificationDispatch, {
        message: <span><Bold>{`${selectedRowsCount} ${selectedRowsCount > 1 ? "entries" : "entry"} `}</Bold>{`${selectedRowsCount > 1 ? "have" : "has"} been added to the reconstructed spec of`} <Bold>{inventoryName}</Bold></span>
    }), [selectedRowsCount, inventoryName, notificationDispatch]);

    const [{loading: submitting, error: submitError}, submitApprovedReview] = useFetch(`apiInventory`, {loadOnMount: false});
    const prevSubmitting = usePrevious(submitting);

    useEffect(() => {
        if (prevSubmitting && !submitting & !submitError) {
            showReviewCompletedNotification();

            returnToInventory();
        }
    }, [prevSubmitting, submitting, submitError, returnToInventory, showReviewCompletedNotification]);
    
    const columns = useMemo(() => [
        {
            Header: 'Path',
            id: "suggestedPath",
            Cell: ({row, data}) => {
                const {original: pathData} = row;
                const {suggestedPath} = pathData;
                const {value, index, id} = openParamInputPathData || {};

                return (
                    <div className="table-row-path-wrapper">
                        <PathDisplay
                            pathData={pathData}
                            isOpenInput={id === row.id}
                            openInputData={{value, index}}
                            onOpenInput={({value, index}) => dispatch({type: REVIEW_ACTIONS.SET_OPEN_PARAM_INPUT_PATH, payload: {id: row.id, value, index}})}
                            onCloseInput={() => dispatch({type: REVIEW_ACTIONS.SET_OPEN_PARAM_INPUT_PATH, payload: {}})}
                            onReviewMerge={({isMerging, paramName}) => {
                                let updatedPath = null;
                                let pathsToReview = [pathData];
                                
                                if (isMerging) {
                                    updatedPath = getPathWithParamInIndex(suggestedPath, index, paramName);

                                    pathsToReview = data.filter(pathData => updatedPath === getPathWithParamInIndex(pathData.suggestedPath, index, paramName));
                                }
                                
                                dispatch({type: REVIEW_ACTIONS.SET_MERGING_PATHS_DATA, payload: {
                                    isMerging,
                                    pathsToReview,
                                    paramPath: updatedPath,
                                    mergeIndex: index,
                                    mergePath: pathData
                                }});
                            }}
                        />
                    </div>
                )
            },
            width: 300
        },
        {
            Header: "Methods",
            id: "methods",
            Cell: ({row}) => {
                const {apiEventsPaths} = row.original;
                const methods = getMethodsFromPaths(apiEventsPaths);
                
                return (
                    <div className="methods-wrapper">{methods.map(method => <Tag key={method}>{method}</Tag>)}</div>
                )
            }
        }
    ], [openParamInputPathData, dispatch]);

    if (!inventoryId) {
        return <Redirect to="/" />;
    }

    if (loading || submitting) {
        return <Loader />;
    }

    if (isLoadingError) {
        return null;
    }

    const {isMerging, pathsToReview, paramPath, mergeIndex, mergePath} = mergingPathsData || {};
    const closeMergingReviewModal = () => dispatch({type: REVIEW_ACTIONS.SET_MERGING_PATHS_DATA, payload: null});

    return (
        <div className="reviewer-page">
            <BackRouteButton title={inventoryName} path={backUrl} query={backQuery} />
            <Title>Spec reviewer</Title>
            <div className="review-actions-wrapper">
                <Button secondary onClick={returnToInventory}>Cancel</Button>
                <Button onClick={() => setShowConfirmationModal(true)} disabled={isEmpty(selectedRowIds)}>Approve review</Button>
            </div>
            <PageContainer>
                <Table
                    columns={columns}
                    withPagination={false}
                    data={{items: dataToReview, total: dataToReview.length}}
                    withMultiSelect={true}
                    onRowSelect={setSelectedRowIds}
                    markedRowIds={!!openParamInputPathData ? [openParamInputPathData.id] : []}
                />
            </PageContainer>
            {!isNull(mergingPathsData) &&
                <MergeModal
                    isMerging={isMerging}
                    paramPath={paramPath}
                    pathsToReview={pathsToReview}
                    mergeIndex={mergeIndex}
                    mergePath={mergePath}
                    onDone={newPaths => {
                        const suggestedPathsToReview = pathsToReview.map(({suggestedPath}) => suggestedPath);
                        const updatedData = [
                            ...newPaths,
                            ...dataToReview.filter(({suggestedPath}) => !suggestedPathsToReview.includes(suggestedPath))
                        ];
                        
                        dispatch({type: REVIEW_ACTIONS.UPDATE_REVIEW_DATA, payload: updatedData});
                    }}
                    onClose={closeMergingReviewModal}
                />
            }
            {showConfirmationModal &&
                <ConfirmationModal
                    inventoryName={inventoryName}
                    pathsCount={selectedRowsCount}
                    onClose={closeConfirmationModal}
                    onConfirm={() => {
                        submitApprovedReview({
                            formatUrl: url => `${url}/${reviewId}/approvedReview`,
                            method: FETCH_METHODS.POST,
                            submitData: {
                                reviewPathItems: dataToReview.filter(item => selectedRowIds.includes(item.id)).map(({id, ...item}) => item)
                            }
                        });

                        closeConfirmationModal();
                    }}
                />
            }
        </div>
    )
}

export default Reviewer;