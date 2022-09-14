import React, { useState, useRef, useEffect } from 'react';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import { useField } from 'formik';
import DropdownButton from 'components/DropdownButton';
import Checkbox from 'components/Checkbox';
import Arrow from 'components/Arrow';
import Tooltip from 'components/Tooltip';
import { FieldLabel, FieldError } from '../utils';

import './multiselect-checkbox-field.scss';

const FieldInput = ({allSelected, value, placeholder, items}) => {

    const getLabel = React.useCallback(() => {
        try {
            const labels = value.map(e => {
                const item = items.find(item => item.value === e);
                return item.label;
            });
            return labels.join(', ');
        } catch(err) {
            return value.join(', ');
        }
    }, [items, value])

    return (
        <div className="input-field">
            {allSelected ? "All" : (isEmpty(value) ? placeholder : getLabel())}
            <Arrow name="bottom" small className="open-menu-icon" />
        </div>
    )
}

const MultiselectCheckboxField = (props) => {
    const {className, label, tooltipText, items, placeholder="Select...", selectAllText = "Select All"} = props;
    const [field, meta, helpers] = useField(props);
    const {name} = field;
    const value = [...field.value];
    const {setValue} = helpers;

    const [isOpen, setIsOpen] = useState(false); 

    const fieldRef = useRef();

    const handleClick = ({target}) => {
        if (fieldRef.current.contains(target)) {
            return;
        }

        setIsOpen(false);
    };

    useEffect(() => {
        if (isOpen) {
            document.addEventListener("mousedown", handleClick);
        } else {
            document.removeEventListener("mousedown", handleClick);
        }
    
        return () => {
            document.removeEventListener("mousedown", handleClick);
        };
      }, [isOpen]);

    const numberSelected = value.length;
    const allSelected = numberSelected === items.length && items.length > 0;
    const halfSelected = numberSelected > 0 && !allSelected;
    const showAllSelected = allSelected || halfSelected;

    const onItemClick = (event) => {
        const {checked, value: clickedValue} = event.target;
        let selectedItems = value;

        if (checked) {
            selectedItems.push(clickedValue);
        } else {
            selectedItems = selectedItems.filter(item => item !== clickedValue);
        }
        
        setValue(selectedItems);
    }

    const onAllClick = (event) => {
        const {checked} = event.target;
        const selectedItems = checked ? items.map(({value}) => value) : [];
        
        setValue(selectedItems);
    }
    
    return (
        <div ref={fieldRef} className={classnames("ps-field-wrapper", "ps-multiselect-checkbox-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <div className="selector-wrapper">
                <DropdownButton
                    toggleButton={<FieldInput allSelected={allSelected} value={value} placeholder={placeholder} items={items}/>}
                    isOpen={isOpen}
                    onToggle={() => setIsOpen(!isOpen)}
                    manualOpen
                >
                    <div className="multiselect-content-container">
                        <Checkbox
                            value="multiselect-show-all"
                            name="multiselect-show-all"
                            title={selectAllText}
                            checked={showAllSelected}
                            onChange={onAllClick}
                            halfSelected={halfSelected}
                            disabled={items.length === 0}
                            small
                        />
                        <div className="multiselect-item-container">
                            {
                                items.map((item,index) => {
                                    if(!item.tooltip){
                                        return (
                                            <Checkbox
                                                key={item.value}
                                                name="multiselect-item"
                                                value={item.value}
                                                title={item.label}
                                                checked={value.includes(item.value)}
                                                onChange={onItemClick}
                                                small
                                            />
                                        )
                                    }
                                    const tooltipId = `api-policy-tooltip-${item.id || index}`;
                                    return (
                                        <div data-tip data-for={tooltipId} key={item.value}>
                                             <Checkbox
                                                name="multiselect-item"
                                                value={item.value}
                                                title={item.label}
                                                checked={value.includes(item.value)}
                                                onChange={onItemClick}
                                                small
                                            />
                                             <Tooltip id={tooltipId} text={item.tooltip} placement="top" />
                                        </div>
                                    )
                                   
                                })
                            }
                            {items.length === 0 && "- no match found -"}
                        </div>
                    </div>
                </DropdownButton>
            </div>
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default MultiselectCheckboxField;