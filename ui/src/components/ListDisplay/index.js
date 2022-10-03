import React from 'react';
import classnames from 'classnames';
import Arrow, { ARROW_NAMES } from 'components/Arrow';

import './list-display.scss';

const ListDisplay = ({items, itemDisplay: ItemDisplay, selectedId, onSelect}) => (
    <div className="list-display-wrapper">
        {
            items.map(item => {
                const {disabled=false} = item;
                return (
                    <div
                        key={item.id}
                        className={classnames("list-display-item-wrapper", { selected: item.id === selectedId }, {disabled})}
                        onClick={disabled ? undefined : () => onSelect(item)}
                    >
                        <div className="list-display-item"><ItemDisplay {...item} /></div>
                        <Arrow name={ARROW_NAMES.RIGHT} small />
                    </div>
                );
            })
        }
    </div>
);

export default ListDisplay;
