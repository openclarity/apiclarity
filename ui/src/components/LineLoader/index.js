import React from 'react';
import classnames from 'classnames';
import { isUndefined } from 'lodash';

import './line-loader.scss';

const LineLoader = ({className, done, total, title, calculatedPercentage, displayItems=false, displayPercent=false, error=false}) => {
    const percentageProvided = !isUndefined(calculatedPercentage);
    const percent = percentageProvided ? calculatedPercentage : done/total*100;

    return (
        <div className={classnames("line-loader-container", {[className]: className})}>
            <div className={classnames("line-loader-filler", {done: percent === 100}, {error: error})} style={{width: `${(done === 0 && !percentageProvided) ? 0.1 : percent}%`}}></div>
            {displayItems && <div className="line-loader-title">{`${done} of ${total} ${title}`}</div>}
            {displayPercent && <div className="line-loader-title">{`${percent}%`}</div>}
        </div>
    );
}

export default LineLoader;