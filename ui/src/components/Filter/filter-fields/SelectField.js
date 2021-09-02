import React, { useEffect, useState, useMemo } from 'react';
import { cloneDeep, isNull, isEqual } from 'lodash';
import classnames from 'classnames';
import { useField } from 'formik';
import DropdownSelect from 'components/DropdownSelect';
import { usePrevious } from 'hooks';

const getMissingValueItemKeys = (valueKey, items) => {
    if (isNull(valueKey)) {
        return items;
    }

    const valueInItems = items.find(item => item.value === valueKey);

    if (!valueInItems) {
        items = cloneDeep(items);
        items.push({value: valueKey, label: valueKey});
    }

    return items;
}

const SelectField = (props) => {
    const {items: fieldItems=[], placeholder, creatable=false, disabled, className} = props;
    const [field, meta, helpers] = useField(props.name);
    const {value} = meta;
    const {name} = field;
    const {setValue} = helpers;
    
    const formattedItems = useMemo(() => (
        creatable && value !== "" ? getMissingValueItemKeys(value, fieldItems) : fieldItems
    ), [creatable, fieldItems, value]);
    const prevFormattedItems = usePrevious(formattedItems);
    
    const [items, setItems] = useState(formattedItems);

    useEffect(() => {
        if (!isEqual(formattedItems, prevFormattedItems)) {
            setItems(formattedItems);
        }
    }, [prevFormattedItems, formattedItems]);

    const selectedValue = items.find(item => item.value === value) || null;
    
    return (
        <DropdownSelect
            className={classnames("filter-form-field", className)}
            name={name}
            value={selectedValue}
            items={items}
            onChange={selectedItem => {
                const {value} = selectedItem;
                
                if (creatable) {
                    setItems(getMissingValueItemKeys(value, items));
                }
                
                setValue(value);
            }}
            creatable={creatable}
            disabled={disabled}
            placeholder={placeholder}
        />
    )
}

export default SelectField;