import React, { useEffect } from 'react';
import { Formik, Form, useFormikContext } from 'formik';
import { isNull, cloneDeep, isEmpty } from 'lodash';
import classnames from 'classnames';
import { TextField, SelectField, MultiselectField, MultiselectCheckboxField, ToggleField, ArrayField,
	RadioField, TextAreaField, ListField, VulnerabilityField, DateField, TimeField, CheckboxField,
	DurationField, AsyncMultiselectField, RadioDescriptionField, YesNoToggleField, FieldLabel, FieldDescription,
	FieldError, KeyValuesWithAllField, ViolationsFilterField } from './form-fields';
import { useFetch, FETCH_METHODS, usePrevious } from 'hooks';
import Loader from 'components/Loader';
import Button from 'components/Button';
import * as validators from './validators';
import * as utils from './utils';

import './form.scss';

const FormComponent = (props) => {
	const {children, className, submitUrl, getSubmitParams, onSubmitSuccess, onSubmitError, onDirtyChanage, hideSubmitButton=false, disableSubmitButton=false,
		saveButtonTitle="Finish", doCustomSubmit, withLoader=true, disableSubmitButtonValidate=false, customSubmitButton: CustomSubmitButton,
		onCancel} = props;
	const {values, isSubmitting, isValidating, setSubmitting, status, setStatus, dirty, isValid, setErrors} = useFormikContext();
	const prevDirty = usePrevious(dirty);

	const [{loading, data, error}, submitFormData] = useFetch(submitUrl, {loadOnMount: false});
	const prevLoading = usePrevious(loading);

	const handleSubmit = () => {
		setSubmitting(true);

        if (!!doCustomSubmit) {
            doCustomSubmit(cloneDeep(values));

            return;
        }

		const submitQueryParams = !!getSubmitParams ? getSubmitParams(cloneDeep(values)) : {};
		submitFormData({method: FETCH_METHODS.POST, submitData: values, ...submitQueryParams});
    }

	useEffect(() => {
		if (prevLoading && !loading) {
			setSubmitting(false);
			setStatus(null);

			if (isNull(error)) {
				if (!!onSubmitSuccess) {
					onSubmitSuccess(data);
				}
			} else {
				const {message, errors} = error;

				if (!!message) {
					setStatus(message);
				}

				if (!isEmpty(errors)) {
					setErrors(errors);
				}

				if (!!onSubmitError) {
					onSubmitError();
				}
			}
		}
	}, [prevLoading, loading, error, data, setSubmitting, setStatus, onSubmitSuccess, setErrors, onSubmitError]);

	useEffect(() => {
		if (prevDirty === dirty) {
			return;
		}

		if (dirty && !!onDirtyChanage) {
			onDirtyChanage(true);
		}
	}, [dirty, prevDirty, onDirtyChanage]);

	if (withLoader && (isSubmitting || loading)) {
		return <Loader />;
	}

	const disableSubmitClick = isSubmitting || isValidating || (!isValid && !disableSubmitButtonValidate) || disableSubmitButton;

	return (
		<Form className={classnames("ps-form", {[className]: className})}>
			{!!status && <div className="main-error-message">{status}</div>}
			{children}
			{!!onCancel && <Button tertiary className="form-cancel-button" onClick={onCancel}>Cancel</Button>}
			{!!CustomSubmitButton && <CustomSubmitButton onClick={handleSubmit} disabled={disableSubmitClick} /> }
			{!hideSubmitButton &&
				<Button
					type="submit"
					className="form-submit-button"
					onClick={handleSubmit}
					disabled={disableSubmitClick}
				>{saveButtonTitle}</Button>
			}
		</Form>
	)
}

const FormWrapper = ({children, initialValues, validate, ...props}) => {
	return (
		<Formik initialValues={initialValues} validate={validate} validateOnMount={true}>
			<FormComponent {...props}>
				{children}
			</FormComponent>
		</Formik>
	)
}

export default FormWrapper;

export {
	useFormikContext,
	validators,
	utils,
	ArrayField,
	TextField,
	SelectField,
	MultiselectField,
	MultiselectCheckboxField,
	ToggleField,
	RadioField,
	TextAreaField,
	ListField,
	VulnerabilityField,
	DateField,
	TimeField,
	CheckboxField,
	DurationField,
	AsyncMultiselectField,
	RadioDescriptionField,
	YesNoToggleField,
	FieldLabel,
	FieldDescription,
	FieldError,
	KeyValuesWithAllField,
	ViolationsFilterField
}
