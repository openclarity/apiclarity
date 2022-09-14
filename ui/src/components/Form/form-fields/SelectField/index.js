import React, { useEffect, useState } from 'react';
import classnames from 'classnames';
import { isEmpty, cloneDeep, isNull } from 'lodash';
import { useField } from 'formik';
import DropdownSelect from 'components/DropdownSelect';
import Loader from 'components/Loader';
import { FieldLabel, FieldError } from '../utils';

import './select-field.scss';

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
    const {items: fieldItems, label, className, placeholder, creatable=false, clearable=true, disabled, tooltipText, loading, small=false} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    let formattedItems = fieldItems || [];
    formattedItems = creatable && value !== "" ? getMissingValueItemKeys(value, formattedItems) : formattedItems;
    const formattedItemsJson = JSON.stringify(formattedItems);
    
    const [items, setItems] = useState(formattedItems);

    useEffect(() => {
        setItems(JSON.parse(formattedItemsJson));
    }, [formattedItemsJson]);
    
    return (
        <div className={classnames("ps-field-wrapper", "ps-select-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <div className="selector-wrapper">
                <DropdownSelect
                    name={name}
                    defaultselected={value}
                    clearable={clearable}
                    items={items}
                    onChange={selectedValue => {
                        if (creatable) {
                            setItems(getMissingValueItemKeys(selectedValue, items));
                        }
                        
                        setValue(selectedValue);
                    }}
                    creatable={creatable}
                    disabled={disabled || loading}
                    placeholder={placeholder}
                    small={small}
                />
                {loading && <Loader small />}
            </div>
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default SelectField;