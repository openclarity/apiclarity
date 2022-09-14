import React, { useEffect, useState } from 'react';
import { isEqual, isNull } from 'lodash';
import { usePrevious } from 'hooks';
import { useFormikContext, validators, SelectField, MultiselectField, ArrayField } from 'components/Form';
import ModalConfirmation from 'components/ModalConfirmation';

const ClearFieldConfirmation = ({allLabel, onCancel, onConfirm}) => (
    <ModalConfirmation
        title="Unsaved changes"
        message={`When choosing '${allLabel}' you are deleting all the data related to this field. Do you want to continue?`}
        confirmTitle="Ok"
        onCancle={onCancel}
        onConfirm={onConfirm}
    />
)

const KeyValuesWithAllField = ({name, label, ALL_KEY, keyName, keyItems, keyPlaceholder, allItemLabel="All", valuesName, getValuesProps, disabled}) => {
    const [prevConfirmFieldValue, setPrevConfirmFieldValue] = useState(null);
    const closeConfirmationModal = () => setPrevConfirmFieldValue(null);

    const {values, setFieldValue} = useFormikContext();

    const fieldValue = values[name];
    const fieldValueJSON = JSON.stringify(fieldValue);
    const prevFieldValueJSON = usePrevious(fieldValueJSON);

    const inUseKeys = fieldValue.map(item => item[keyName]).filter(item => item !== "");
    const formattedKeyItems = [{value: ALL_KEY, label: allItemLabel}, ...keyItems]
        .map(item => ({...item, isDisabled: inUseKeys.includes(item.value)}));

    useEffect(() => {
        if (fieldValueJSON === prevFieldValueJSON || !prevFieldValueJSON) {
            return;
        }

        const prevFieldValue = JSON.parse(prevFieldValueJSON);
        const formattedFieldValue = JSON.parse(fieldValueJSON).map((item, index, items) => {
            const keyValue = item[keyName];
            const valuesValue = item[valuesName];

            if (keyValue === ALL_KEY && items.length > 1) {
                setPrevConfirmFieldValue(prevFieldValue);

                return item;
            }

            if (keyValue === ALL_KEY && valuesValue.length > 0) {
                return {[keyName]: keyValue, [valuesName]: []}
            }

            const prevKeyItem = prevFieldValue.find(item => item[keyName] === keyValue);
            const prevValuesValue = !!prevKeyItem ? prevKeyItem[valuesName] : [];

            if (isEqual(prevValuesValue, valuesValue)) {
                return item;
            }
            
            const prevHasAll = !!prevValuesValue.find(item => item === ALL_KEY);
            const currHasAll = !!valuesValue.find(item => item === ALL_KEY);

            let updatedItem = item;
            if (valuesValue.length > 1) {
                if (prevHasAll) {
                    //new non All item was added => remove the ALL item:
                    const nonAllvalues = valuesValue.filter(name => name !== ALL_KEY);
                    updatedItem = {...updatedItem, [valuesName]: nonAllvalues};
    
                } else if (currHasAll) {
                    //All item was added => remove the non ALL items:
                    updatedItem = {...updatedItem, [valuesName]: [ALL_KEY]};
                }
            }
            return updatedItem;
        });
        setFieldValue(name, formattedFieldValue);

    }, [fieldValueJSON, prevFieldValueJSON, setFieldValue, name, ALL_KEY, keyName, valuesName]);

    return (
        <React.Fragment>
            <ArrayField
                name={name}
                label={label}
                firstFieldProps={{
                    component: SelectField,
                    key: keyName,
                    placeholder: keyPlaceholder,
                    items: formattedKeyItems,
                    clearable: false
                }}
                secondFieldProps={{
                    component: MultiselectField,
                    key: valuesName,
                    emptyValue: [],
                    items: [],
                    getDependentFieldProps: (data) => {
                        const keyValue =  data[keyName];

                        return {
                            disabled: keyValue === ALL_KEY || keyValue === "",
                            validate: keyValue === ALL_KEY ? undefined : validators.validateRequired,
                            ...getValuesProps(data)
                        }
                    }
                }}
                disabled={disabled}
            />
            {!isNull(prevConfirmFieldValue) &&
                <ClearFieldConfirmation
                    allLabel={allItemLabel}
                    onConfirm={() => {
                        setFieldValue(name, [{[keyName]: ALL_KEY, [valuesName]: []}]);
                        closeConfirmationModal();
                    }}
                    onCancel={() => {
                        setFieldValue(name, prevConfirmFieldValue);
                        closeConfirmationModal();
                    }}
                />
            }
        </React.Fragment>
    );
}

export default KeyValuesWithAllField;