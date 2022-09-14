import React from 'react';
import { isEmpty } from 'lodash';
import { Field } from 'formik';
import { FieldLabel, FieldError } from '../utils';

import './text-area-field.scss';

const TextAreaField = ({name, label, tooltipText, placeholder, disabled, validate}) => (
	<Field name={name} validate={validate}>
		{({field, meta}) => {
			return (
				<div className="ps-field-wrapper ps-text-area-wrapper">
					{!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
					<textarea {...field} placeholder={placeholder} disabled={disabled} />
					{meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
				</div>
			)
		}}
	</Field>
)

export default TextAreaField;