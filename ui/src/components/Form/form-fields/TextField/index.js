import React, {useState} from 'react';
import { isEmpty } from 'lodash';
import classnames from 'classnames';
import { Field } from 'formik';
import Icon, { ICON_NAMES } from 'components/Icon';
import { FieldLabel, FieldError } from '../utils';

import './text-field.scss';

const TYPE_TEXT = "text";
const TYPE_PASSWORD = "password";

const TextField = ({name, type=TYPE_TEXT, label, tooltipText, placeholder, disabled, withPasswordShow=false, validate, fieldUnits}) => {
	const [inputType, setInputType] = useState(type);
	const hasLabel = !!label;

	const onShowPasswordClick = () => {
        if (!withPasswordShow) {
            return;
		}

        setInputType(inputType === TYPE_PASSWORD ? TYPE_TEXT : TYPE_PASSWORD)
	}
	
	return (
		<Field name={name} validate={validate}>
			{({field, meta}) => {
				return (
					<div className={classnames("ps-field-wrapper", "ps-text-field-wrapper", {"has-label": hasLabel})}>
						{hasLabel && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
						<input {...field} type={inputType} placeholder={placeholder} disabled={disabled} />
						{(!isEmpty(fieldUnits) && !disabled) && <div className="field-units">{fieldUnits}</div>}
						{(withPasswordShow && !disabled) &&
							<Icon className="password-icon" name={inputType === TYPE_PASSWORD ? ICON_NAMES.EYE : ICON_NAMES.EYE_HIDE} onClick={onShowPasswordClick} />}
						{meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
					</div>
				)
			}}
		</Field>
	);
}

export default TextField;