import React, { useState } from 'react';
import { isEmpty } from 'lodash';
import Modal from 'components/Modal';
import MethodTag from 'components/MethodTag';
import CheckboxListSelect from 'components/CheckboxListSelect';
import { SEPARATOR, checkIsParam, getMethodsFromPaths } from '../utils';

import './merge-modal.scss';

const MergeIntro = ({pathsToReview, paramName}) => {
    const pathsCount = pathsToReview.length;

    return (
        <React.Fragment>
            {`We selected ${pathsCount} path${pathsCount>1 ? "s" : ""} that can be shortened using the parameter `}
            <span style={{fontWeight: "bold"}}>{paramName}</span>
        </React.Fragment>
    );
}

const UnmergeIntro = ({pathsToReview, usedParamNames}) => {
    const pathsCount = pathsToReview.reduce((acc, curr) => {
        return acc + curr.apiEventsPaths.length;
    }, 0);

    return (
        <React.Fragment>
            {`We selected ${pathsCount} path${pathsCount>1 ? "s" : ""} that have been shortened and merged using the parameter${usedParamNames.length>1 ? "s" : ""} `}
            {
                usedParamNames.map((paramName, index, paramNames) => (
                    <span key={index}><span style={{fontWeight: "bold"}}>{paramName}</span>{index < paramNames.length -1  ? " and " : ""}</span>
                ))
            }
            :
        </React.Fragment>
    );
}

const ReviewPathDisplay = ({path, methods, markIndexs}) => (
    <div className="merge-review-path-display">
        <div className="merge-review-path-display-path">
            {
                path.split(SEPARATOR).map((section, index, pathList) => (
                    <span key={index} style={{fontWeight: markIndexs.includes(index) || checkIsParam(section) ? "bold" : "normal"}}>
                        {section}{index < pathList.length -1  ? SEPARATOR : ""}
                    </span>
                ))
            }
        </div>
        <div className="merge-review-path-display-methods">{methods.map(method => <MethodTag key={method} method={method} />)}</div>
    </div>
);

const MergeReviewItems = ({pathsToReview, selectedItems, setSelectedItems, markIndexs}) => (
    <CheckboxListSelect
        items={pathsToReview}
        titleDisplay={({suggestedPath, apiEventsPaths}) => (
            <ReviewPathDisplay path={suggestedPath} methods={getMethodsFromPaths(apiEventsPaths)} markIndexs={markIndexs} />
        )}
        selectedItems={selectedItems}
        setSelectedItems={setSelectedItems}
    />
);

const UnmergeReviewItems = ({pathsToReview, selectedItems, setSelectedItems, markIndexs}) => (
    <React.Fragment>
        {
            pathsToReview.map(({apiEventsPaths}, index) => (
                <React.Fragment key={index}>
                    {
                        <CheckboxListSelect
                            items={apiEventsPaths.map(item => ({...item, id: item.path}))}
                            titleDisplay={({path, methods}) => <ReviewPathDisplay path={path} methods={methods} markIndexs={markIndexs} />}
                            selectedItems={selectedItems}
                            setSelectedItems={setSelectedItems}
                        />
                    }
                </React.Fragment>
            ))
        }
    </React.Fragment>
);

const MergeModal = ({isMerging=true, pathsToReview, paramPath, mergeIndex, mergePath, onClose, onDone}) => {
    const [selectedItems, setSelectedItems] = useState(isMerging ? [mergePath] : []);

    const ReviewItems = isMerging ? MergeReviewItems : UnmergeReviewItems;
    const usedParamNamesWithIndex = pathsToReview[0].suggestedPath.split(SEPARATOR)
        .map((item, index) => ({section: item, index})).filter(({section}) => checkIsParam(section));
    
    const onMerge = () => {
        const selectedSuggestedPaths = selectedItems.map(({suggestedPath}) => suggestedPath);
        const selectedApiPathItems = selectedItems.reduce((acc, {apiEventsPaths}) => {
            return [...acc, ...apiEventsPaths];
        }, []);
        const notSelectedPaths = pathsToReview.filter(({suggestedPath}) => !selectedSuggestedPaths.includes(suggestedPath));

        const newPaths = [
            {suggestedPath: paramPath, apiEventsPaths: selectedApiPathItems},
            ...notSelectedPaths
        ];
        
        onDone(newPaths);
    }

    const onUnmerge = () => {
        const apiEventPathsToUnmerge = selectedItems.map(({path}) => path);
        const {suggestedPath, apiEventsPaths} = pathsToReview[0];
        const apiEventsPathsLeft = apiEventsPaths.filter(({path}) => !apiEventPathsToUnmerge.includes(path));

        const newPaths = [
            ...(isEmpty(apiEventsPathsLeft) ? [] : [{suggestedPath, apiEventsPaths: apiEventsPathsLeft}]),
            ...selectedItems.map(({path, methods}) => ({suggestedPath: path, apiEventsPaths: [{path, methods}]}))
        ]
        
        onDone(newPaths);
    }
    
    return (
        <Modal
            title={isMerging ? "Merge entries" : "Restore and Unmerge"}
            className="merge-modal"
            onClose={onClose}
            onDone={isMerging ? onMerge : onUnmerge} 
            doneTitle={isMerging ? "Merge and shorten" : "Restore and unmerge"}
            disableDone={isEmpty(selectedItems)}
        >
            <div>
                {isMerging ? <MergeIntro pathsToReview={pathsToReview} paramName={paramPath.split(SEPARATOR)[mergeIndex]} /> :
                    <UnmergeIntro pathsToReview={pathsToReview} usedParamNames={usedParamNamesWithIndex.map(({section}) => section)} />}
            </div>
            <div className="select-wrapper">
                <ReviewItems
                    pathsToReview={pathsToReview}
                    selectedItems={selectedItems}
                    setSelectedItems={setSelectedItems}
                    markIndexs={isMerging ? [mergeIndex] : usedParamNamesWithIndex.map(({index}) => index)}
                /> 
            </div>
            <div>{`Select the entries you want to ${isMerging ? "merge and shorten" : "restore and unmerge"}.`}</div>
        </Modal>
    );
}

export default MergeModal;