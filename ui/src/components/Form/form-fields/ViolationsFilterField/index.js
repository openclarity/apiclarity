import React, { useState, useRef, useEffect } from 'react';
import classnames from 'classnames';
import { isEmpty, cloneDeep, debounce } from 'lodash';
import { useField } from 'formik';
import DropdownButton from 'components/DropdownButton';
import Checkbox from 'components/Checkbox';
import Arrow from 'components/Arrow';
import ToggleButton from 'components/ToggleButton';
import { FieldLabel, FieldError } from 'components/Form';
import { usePrevious } from 'hooks';
import { RULE_ACTIONS, RISKY_ACTION_ITEM } from 'utils/systemConsts';

import './violations-filter-field.scss';

const VIOLATION_RULE_ACTIONS = [RULE_ACTIONS.DETECT, RULE_ACTIONS.BLOCK];

const Indicator = ({className, selected}) => <div className={classnames("indicator-item", className, {selected})}></div>;
const SelectIndicator = ({value, withEncrypt, withRisky}) => (
    <div className="select-indicator">
        <div className="indicator-container">
            <Indicator className={RULE_ACTIONS.ALLOW.toLowerCase()} selected={isEmpty(value) || value.includes(RULE_ACTIONS.ALLOW)} />
            {withEncrypt && 
                <Indicator className={RULE_ACTIONS.ENCRYPT.toLowerCase()} selected={isEmpty(value) || value.includes(RULE_ACTIONS.ENCRYPT)} />
            }
            {withRisky && 
                <Indicator className={RISKY_ACTION_ITEM.value.toLowerCase()} selected={isEmpty(value) || value.includes(RISKY_ACTION_ITEM.value)} />
            }
            <Indicator className={RULE_ACTIONS.DETECT.toLowerCase()} selected={isEmpty(value) || value.includes(RULE_ACTIONS.DETECT)} />
            <Indicator className={RULE_ACTIONS.BLOCK.toLowerCase()} selected={isEmpty(value) || value.includes(RULE_ACTIONS.BLOCK)} />
        </div>
        <Arrow name="bottom" className="open-menu-icon" small />
    </div>
);

const ViolationsFilterField = (props) => {
    const {className, label, items} = props;
    const [field, meta, helpers] = useField(props);
    const {value} = field; 
    const {setValue} = helpers;
    const [isOpen, setIsOpen] = useState(false);
    const [violationsOnly, setViolationsOnly] = useState(false);
    const prevViolationsOnly = usePrevious(violationsOnly);
    const [fieldItems, setFieldItems] = useState(items);

    const fieldRef = useRef();
    const inititalLoaded = useRef(false);

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

    useEffect(() => {
        if (!inititalLoaded.current) {
            inititalLoaded.current = true;

            return;
        }

        if (prevViolationsOnly === violationsOnly) {
            return;
        }

        let formattedItems = cloneDeep(items);

        if (violationsOnly) {
            formattedItems.map(item => {
                if (!VIOLATION_RULE_ACTIONS.includes(item.value)) {
                    item.disabled = true;
                }

                return item;
            });
        } else {
            formattedItems = formattedItems.map(({value, label}) => ({value, label}));
        }

        setFieldItems(formattedItems);
        debounce(setValue, 150)(violationsOnly ? VIOLATION_RULE_ACTIONS : []);

    }, [violationsOnly, prevViolationsOnly, setValue, items]);

    const onItemClick = (event) => {
        const {checked, value: clickedValue} = event.target;
        let selectedItems = [...value];

        if (checked) {
            selectedItems.push(clickedValue);
        } else {
            selectedItems = selectedItems.filter(item => item !== clickedValue);
        }
        
        setValue(selectedItems);
    }

    const withEncrypt = !!items.find(item => item.value === RULE_ACTIONS.ENCRYPT);
    const withRisky = !!items.find(item => item.value === RISKY_ACTION_ITEM.value);
    
    return (
        <div ref={fieldRef} className={classnames("violations-filter-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel>{label}</FieldLabel>}
            <div className="selector-wrapper">
                <ToggleButton
                    className="vulnerabilities-ignore-toggle"
                    checked={violationsOnly}
                    onChange={() => setViolationsOnly(!violationsOnly)}
                />
                <DropdownButton
                    toggleButton={<SelectIndicator value={value} withEncrypt={withEncrypt} withRisky={withRisky} />}
                    isOpen={isOpen}
                    onToggle={() => setIsOpen(!isOpen)}
                    manualOpen
                >
                    <div className="multiselect-content-container">
                        {
                            fieldItems.map(item => (
                                <Checkbox
                                    key={item.value}
                                    name="multiselect-item"
                                    value={item.value}
                                    title={item.label}
                                    checked={value.includes(item.value)}
                                    onChange={onItemClick}
                                    disabled={item.disabled}
                                    small
                                />
                            ))
                        }
                        {fieldItems.length === 0 && "- no match found -"}
                    </div>
                </DropdownButton>
            </div>
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default ViolationsFilterField;