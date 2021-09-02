import React, { useState, useEffect } from 'react';
import Checkbox from 'components/Checkbox';

import './checkbox-list-select.scss';

const CheckboxListSelect = ({items, titleDisplay: TitleDisplay, selectedItems, setSelectedItems}) => {
    const [selectAllChecked, setSelectAllChecked] = useState(false);

    const noSelectItems = items.length === 0;
    const allItemsSelected = selectedItems.length === items.length;
    const noItemsSelected = selectedItems.length === 0;
    
    useEffect(() => {
        if (noSelectItems) {
            return;
        }

        if (allItemsSelected && !selectAllChecked) {
            setSelectAllChecked(true);
        } else if (noItemsSelected && selectAllChecked) {
            setSelectAllChecked(false);
        }
    }, [noSelectItems, allItemsSelected, noItemsSelected, selectAllChecked, setSelectAllChecked]);

    const onSelectAll = (checked) => setSelectedItems(checked ? items : []);

    const onItemSelect = (item, checked) => setSelectedItems(checked ? [...selectedItems, item] : selectedItems.filter(({id}) => id !== item.id));

    return (
        <div className="checkbox-list-select-wrapper">
            <Checkbox title="Select / Deselect all" checked={selectAllChecked} halfSelected={!allItemsSelected} onChange={event => onSelectAll(event.target.checked)} />
            {
                items.map((item, index) => (
                    <Checkbox
                        key={index}
                        title={<TitleDisplay {...item} />}
                        checked={!!selectedItems.find(({id}) => id === item.id)}
                        onChange={(event) => onItemSelect(item, event.target.checked)}
                    />
                ))
            }
            {noSelectItems && <div>--- no items to select ---</div>}
        </div>
    )
}

export default CheckboxListSelect;