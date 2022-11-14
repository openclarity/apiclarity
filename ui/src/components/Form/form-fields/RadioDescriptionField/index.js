import React, { useState } from 'react';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import { useField } from 'formik';
import Arrow from 'components/Arrow';
import { FieldLabel, FieldError } from '../utils';

import './radio-description-field.scss';

const RadioButtons = ({name, items, selected, onChange, disabled}) => {
    const [openItem, setOpenItem] = useState(null);

    return (
        <div className={classnames("ps-radio-desc-container")}>
            {
                items.map(({value, label, description}) => {
                    const isSelected = selected === value;
                    const isOpenSet = openItem === value;
                    const isOpen = isOpenSet || isSelected;
                    
                    return (
                        <div key={value} className={classnames("radio-item-wrapper", {disabled})}>
                            <div className="radio-item-title">
                                <label className="ps-radio">
                                    <span className="ps-radio-text">{label}</span>
                                    <input
                                        type="radio"
                                        name={name}
                                        checked={isSelected}
                                        value={value}
                                        onChange={() => onChange(value)}
                                        disabled={disabled}
                                    />
                                    <span className="checkmark"></span>
                                </label>
                                <Arrow
                                    name={isOpen ? "top" : "bottom"}
                                    small
                                    className="arrow-icon"
                                    onClick={() => setOpenItem(isOpenSet ? null : value)}
                                />
                            </div>
                            {isOpen && <div className="radio-item-description">{description}</div>}
                        </div>
                    )
                })
            }
        </div>
    );
}

const RadioDescriptionField = (props) => {
    const {items, label, className, tooltipText, disabled} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    return (
        <div className={classnames("ps-field-wrapper", "ps-radio-description-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <RadioButtons
                name={name}
                items={items}
                selected={value}
                onChange={value => setValue(value)}
                disabled={disabled}
            />
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default RadioDescriptionField;