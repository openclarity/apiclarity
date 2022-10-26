import React from 'react'
import classNames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';

import './list-item-display-bfla.scss'

const ListItemDisplayBFLA = ({ name, namespace, isLegitimate, method }) => {
    return (
        <div
            style={{
                display: 'flex',
                justifyContent: 'space-between'
            }}>
            <div className='display-list-item-container'>
                {
                    method &&
                    <div className='display-list-item-name'>
                        {method}
                    </div>
                }
                <div className={method ? 'display-list-item-path' : 'display-list-item-name'}>
                    {name}
                </div>
                {namespace && (
                    <div className='display-list-item-namespace'>
                        &nbsp;|&nbsp;
                        {namespace}
                    </div>
                )}
            </div>
            <Icon
                className={classNames('bfla-status-icon', isLegitimate ? "legitimate" : "illegitimate")}
                name={
                    isLegitimate ?
                        ICON_NAMES.SHIELD_CHECK :
                        ICON_NAMES.SHIELD_CROSS
                }
            />
        </div>
    )
}

export default ListItemDisplayBFLA
