import React from 'react';
import Icon from 'components/Icon';

import './key-value-list.scss';

const KeyValueItem = ({label, value, icon, hideEmpty=false}) => (
    <div className="key-value-item-wrapper">
        <div className="key-wrapper">{label}</div>
        <div className="value-container">
            <div className="value-wrapper">{!!value || value === 0 ? value : (hideEmpty ? "" : 'N/A')}</div>
            {!!icon && <Icon name={icon} className="value-icon"/>}
        </div>
    </div>
)

const KeyValueList = ({items}) => (
    <div className="key-value-list-wrapper">
        {items.map((item, index) => <KeyValueItem key={index} {...item} />)}
    </div>
);

export default KeyValueList;
