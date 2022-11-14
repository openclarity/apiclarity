import React from 'react';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import { useField } from 'formik';
import AsyncSelect from 'react-select/async';
import AsyncCreatableSelect from 'react-select/async-creatable';
import { dataMultiFetcher, asyncDataFetcher } from 'utils/apiUtils';
import { useNotificationDispatch, showNotification, NOTIFICATION_TYPES } from 'context/NotificationProvider';
import { FieldLabel, FieldError } from '../utils';

import './async-multiselect-field.scss';

const AsyncMultiselectField = (props) => {
    const {className, tooltipText, getUrlsData, formatData, url, getQueryParams, label, creatable, placeholder="Select...", disabled} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    const notificationDispatch = useNotificationDispatch();
    const showErrorToaster = () => showNotification(notificationDispatch, {message: "An error occurred when trying to load data", type: NOTIFICATION_TYPES.ERROR});

    const promiseOptions = inputValue => {
        const isSingleFetch = !!url;
        const fetcher = isSingleFetch ? asyncDataFetcher : dataMultiFetcher;
        
        return new Promise (resolve => fetcher({
            url, // asyncDataFetcher
            queryParams: isSingleFetch ? getQueryParams(inputValue) : null, //asyncDataFetcher
            urlsData: isSingleFetch ? null : getUrlsData(inputValue), //dataMultiFetcher
            successCallback: (data) => {
                resolve(formatData(data));
            },
            errorCallback: (error) => {
                showErrorToaster();

                resolve([]);
            },
            authenticationErrorCallback: () => {
                showErrorToaster();

                resolve([]);
            }
        }))
    }

    const SelectComponent = creatable ? AsyncCreatableSelect : AsyncSelect;
    const selectProps = {
        value,
        isMulti: true,
        name,
        className: "ps-multi-select",
        classNamePrefix: "multi-select",
        onChange: selectedItems => setValue(selectedItems),
        loadOptions: promiseOptions,
        placeholder,
        noOptionsMessage: ({inputValue}) => {
            if (inputValue === "") {
                return placeholder;
            }
            
            return "No options";
        },
        isDisabled: disabled
    };

    return (
        <div className={classnames("ps-field-wrapper", "ps-multiselect-field-wrapper", "is-async", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <div className="selector-wrapper">
                <SelectComponent {...selectProps} />
            </div>
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default AsyncMultiselectField;