import React from 'react';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import { useField } from 'formik';
import RadioButtons from 'components/RadioButtons';
import { FieldLabel, FieldError } from '../utils';

const RadioField = (props) => {
    const {items, label, className, tooltipText, horizontal=false, disabled=false} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    return (
        <div className={classnames("ps-field-wrapper", "ps-radio-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <RadioButtons
                name={name}
                items={items}
                selected={value}
                onChange={value => setValue(value)}
                horizontal={horizontal}
                disabled={disabled}
            />
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default RadioField;