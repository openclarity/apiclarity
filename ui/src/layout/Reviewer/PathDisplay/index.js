import React, { useState } from 'react';
import classnames from 'classnames';
import Button from 'components/Button';
import Icon, { ICON_NAMES } from 'components/Icon';
import Tooltip from 'components/Tooltip';
import { SEPARATOR, checkIsParam } from '../utils';

import './path-display.scss';

const ParamInputDisplay = ({onClose, onDone, paramValue, isParam}) => {
    const initialValue = isParam ? paramValue.replace(/{|}/g, "") : "";
    const [paramName, setParamName] = useState(initialValue);
    
    return (
        <div className="param-input-wrapper">
            <div className="param-input-title">{isParam ? "Update parameter name:" : "Parameter name:"}</div>
            <div className="param-input-container">
                <input
                    type="text"
                    placeholder={isParam ? `replace ${initialValue} with...` : `"${paramValue}" is an example of a...`}
                    value={paramName}
                    onChange={event => setParamName(event.target.value)}
                />
                <Button
                    className="param-ok-button"
                    onClick={() => onDone(paramName)}
                    disabled={paramName === "" || paramName.includes("{") || paramName.includes("}")}
                >OK</Button>
                <Button secondary onClick={onClose}>Cancel</Button>
            </div>
        </div>
    )
}

const UmergeIcon = ({tooltipId, onClick}) => (
    <div className="umerge-icon-wrapper">
        <div data-tip data-for={tooltipId}>
            <Icon name={ICON_NAMES.UNMERGE} onClick={onClick} />
        </div>
        <Tooltip id={tooltipId} text="Restore and unmerge" />
    </div>
);

const PathDisplay = ({pathData, isOpenInput, openInputData, onOpenInput, onCloseInput, onReviewMerge, onRenameParam}) => {
    const {id, suggestedPath} = pathData;
    const pathList = suggestedPath.split(SEPARATOR).map(section => ({value: section, isParam: checkIsParam(section)}));
    const pathHasParam = !!pathList.find(({isParam}) => isParam);

    const onSectionClick = ({value, index, isParam}) => {
        if (isOpenInput) {
            return;
        }

        onOpenInput({value, index, id, isParam});
    }
    
    return (
        <div className={classnames("path-container", {open: isOpenInput})}>
            <div className="path-display-wrapper">
                <div className="path-wrapper">
                    {
                        pathList.map(({value, isParam}, index, pathList) => (
                            <span key={index}>
                                <span
                                    className={classnames("path-item", {param: isParam})}
                                    onClick={() => onSectionClick({value, index, isParam})}
                                >{value}</span>
                                {index < pathList.length -1  ? SEPARATOR : ""}
                            </span>
                        ))
                    }
                    
                </div>
                {pathHasParam && <UmergeIcon tooltipId={`unmerge-button-tooltip-${id}`} onClick={() => onReviewMerge({isMerging: false})} />}
            </div>
            {isOpenInput &&
                <ParamInputDisplay
                    onClose={onCloseInput}
                    onDone={paramName => {
                        if (openInputData.isParam) {
                            onRenameParam({paramName})
                        } else {
                            onReviewMerge({isMerging: true, paramName})
                        }
                    }}
                    paramValue={openInputData.value}
                    isParam={openInputData.isParam}
                />
            }
        </div>
    );
}

export default PathDisplay;