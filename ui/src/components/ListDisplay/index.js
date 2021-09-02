import React from 'react';
import classnames from 'classnames';
import Arrow, { ARROW_NAMES } from 'components/Arrow';

import './list-display.scss';

const ListDisplay = ({items, itemDisplay: ItemDisplay, selectedId, onSelect}) => (
    <div className="list-display-wrapper">
        {
            items.map(item => (
                <div
                    key={item.id}
                    className={classnames("list-display-item-wrapper", {selected: item.id === selectedId})}
                    onClick={() => onSelect(item)}
                >
                    <div className="list-display-item"><ItemDisplay {...item} /></div>
                    <Arrow name={ARROW_NAMES.RIGHT} small />
                </div>
            ))
        }
    </div>
);

export default ListDisplay;