import React from 'react'
import classNames from 'classnames'
import Arrow, { ARROW_NAMES } from 'components/Arrow';

export default function ListDisplayItem({
    title,
    isSelected,
    icon,
    onSelect,
    disabled = false,
    index,
}) {

    return (
        <div
            className={
                classNames(
                    'list-display-item-container',
                    isSelected && 'selected'
                )
            }
            onClick={!disabled ? () => onSelect(index) : undefined}
        >
            <div className='list-display-item-title'>
                {title}
            </div>
            <div style={{
                display: 'flex'
            }}>
                <div className='list-display-item-icon'>
                    {icon}
                </div>
                <div className='list-display-item-arrow-container'>
                    {!disabled && <Arrow name={isSelected ? ARROW_NAMES.LEFT : ARROW_NAMES.RIGHT} small />}
                </div>
            </div>
        </div>
    )
}
