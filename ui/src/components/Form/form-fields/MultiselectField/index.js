import React, { useState, useEffect } from 'react';
import classnames from 'classnames';
import { isEmpty, cloneDeep } from 'lodash';
import { useField } from 'formik';
import Select from 'react-select';
import CreatableSelect from 'react-select/creatable';
import Loader from 'components/Loader';
import Tooltip from 'components/Tooltip';
import { FieldLabel, FieldError } from '../utils';

import './multiselect-field.scss';

const getMissingValueItemKeys = (valueKeys, items) => {
    const missingItems = valueKeys.filter(key => !items.find(item => item.value === key));

    if(missingItems.length > 0) {
        items = cloneDeep(items);
        missingItems.forEach(item => {
            items.push({value: item, label: item});
        });
    }

    return items;
}

const MultiselectField = (props) => {
    const {items: fieldItems, label, className, placeholder="Select...", creatable=false, clearable=true, disabled, tooltipText,
        loading=false, showValueTooltip=false, components, fullObjectValue=false} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    let formattedItems = fieldItems || [];
    formattedItems = creatable ? getMissingValueItemKeys(value, formattedItems) : formattedItems;
    const formattedItemsJson = JSON.stringify(formattedItems);

    const [items, setItems] = useState(formattedItems);

    useEffect(() => {
        setItems(JSON.parse(formattedItemsJson));
    }, [formattedItemsJson]);

    const selectedItems = fullObjectValue ? value : items.filter(item => value.includes(item.value));
    const SelectComponent = creatable ? CreatableSelect : Select;

    const valueTooltipId = `value-tooltip-id-${name}`;
    
    return (
        <div className={classnames("ps-field-wrapper", "ps-multiselect-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <div className="selector-wrapper" data-tip data-for={valueTooltipId}>
                <SelectComponent
                    value={selectedItems}
                    isMulti
                    name={name}
                    options={items}
                    className="ps-multi-select"
                    classNamePrefix="multi-select"
                    onChange={selectedItems => {
                        const formattedSelectedItems = selectedItems || [];
                        const valueKeys = formattedSelectedItems.map(item => item.value);

                        if (creatable) {
                            setItems(getMissingValueItemKeys(valueKeys, items));
                        }
                        
                        setValue(fullObjectValue ? formattedSelectedItems : valueKeys);
                    }}
                    isDisabled={disabled || loading}
                    isClearable={clearable}
                    placeholder={placeholder}
                    components={components}
                />
                {loading && <Loader small />}
            </div>
            {showValueTooltip && !isEmpty(selectedItems) &&
                <Tooltip id={valueTooltipId} text={selectedItems.map(item => item.label).join(", ")} placement="top" />}
            {meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default MultiselectField;