import React from 'react';
import { isEmpty } from 'lodash';
import classnames from 'classnames';
import TimeField from 'react-simple-timefield';
import { useField } from 'formik';
import { FieldLabel, FieldError } from '../utils';

import './time-field.scss';

const FormTimeField = (props) => {
    const {label, className, tooltipText} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    return (
        <div className={classnames("ps-field-wrapper", "ps-time-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <div className="selector-wrapper">
                <TimeField className="time-input" value={value} onChange={(event, value) => setValue(value)} />
            </div>
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default FormTimeField;