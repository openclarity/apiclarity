import React, { useState } from 'react';
import classnames from 'classnames';
import Button from 'components/Button';
import { SEPARATOR, checkIsParam } from '../utils';

import './path-display.scss';

const ParamInputDisplay = ({onClose, onDone, paramValue}) => {
    const [paramName, setParamName] = useState("");

    return (
        <div className="param-input-wrapper">
            <div className="param-input-title">Parameter name:</div>
            <div className="param-input-container">
                <input
                    type="text"
                    placeholder={`"${paramValue}" is an example of a...`}
                    value={paramName}
                    onChange={event => setParamName(event.target.value)}
                />
                <Button className="param-ok-button" onClick={() => onDone(paramName)} disabled={paramName === ""}>OK</Button>
                <Button secondary onClick={onClose}>Cancel</Button>
            </div>
        </div>
    )
}

const PathDisplay = ({pathData, isOpenInput, openInputData, onOpenInput, onCloseInput, onReviewMerge}) => {
    const {id, suggestedPath} = pathData;
    const pathList = suggestedPath.split(SEPARATOR).map(section => ({value: section, isParam: checkIsParam(section)}));

    const onSectionClick = ({value, index, isParam}) => {
        if (isOpenInput) {
            return;
        }

        if (isParam) {
            onReviewMerge({isMerging: false});
        } else {
            onOpenInput({value, index, id});
        }
    }
    
    return (
        <React.Fragment>
            <div className={classnames("path-display-wrapper", {open: isOpenInput})}>
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
                {isOpenInput &&
                    <ParamInputDisplay
                        onClose={onCloseInput}
                        onDone={paramName => onReviewMerge({isMerging: true, paramName})}
                        paramValue={openInputData.value}
                    />
                }
            </div>
        </React.Fragment>
    );
}

export default PathDisplay;