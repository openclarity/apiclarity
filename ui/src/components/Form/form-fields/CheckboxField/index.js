import React from 'react';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import { useField } from 'formik';
import Checkbox from 'components/Checkbox';
import { FieldLabel, FieldError } from '../utils';

const CheckboxField = (props) => {
    const {title, label, className, tooltipText, disabled=false} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    const formattedValue = value || false;

    return (
        <div className={classnames("ps-field-wrapper", "ps-checkbox-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <Checkbox
                value={formattedValue}
                name={name}
                title={title}
                checked={formattedValue}
                onChange={event => setValue(event.target.checked)}
                disabled={disabled}
            />
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default CheckboxField;