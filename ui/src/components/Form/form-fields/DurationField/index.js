import React from 'react';
import classnames from 'classnames';
import { useField } from 'formik';
import { isEmpty } from 'lodash';
import { FieldLabel, FieldError } from '../utils';

import './duration-field.scss';

const DurationField = (props) => {
    const {label, placeholder, disabled=false, duration, durationPrefix, min, max, className, tooltipText} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    return (
        <div className={classnames("ps-field-wrapper", "ps-duration-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <div className="duration-input-container">
                {!!durationPrefix && <div className="duration-prefix">{durationPrefix}</div>}
                <input
                    name={name}
                    value={value}
                    min={min}
                    max={max}
                    type="number"
                    placeholder={placeholder}
                    disabled={disabled}
                    onChange={event => setValue(event.target.value)}
                />
                <div className="duration-type">{duration}</div>
            </div>
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default DurationField;