import React, { useEffect } from 'react';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import { FieldArray, useField, useFormikContext } from 'formik';
import RoundIconContainer from 'components/RoundIconContainer';
import { FieldLabel } from '../utils';

import './array-field.scss';

const FieldItemWrapper = ({name, value, index, push, remove, firstFieldProps, secondFieldProps, disabled, horizontal=true, withBackground=false}) => {
    const {component: FirstFieldComponent, emptyValue: firstEmptyValue="", ...firstProps} = firstFieldProps;
    const {component: SecondFieldComponent, getDependentFieldProps, emptyValue: secondEmptyValue="", ...secondProps} = secondFieldProps;

    const firstKey = firstProps.key;
    const secondKey = secondProps.key;

    const allowRemove = value.length > 1;

    const formattedFirstProps = {
        ...firstProps,
        disabled
    };
    const formattedSecondProps = {
        ...secondProps,
        ...(!getDependentFieldProps ? {} : {...getDependentFieldProps(value[index]), index})
    };
    formattedSecondProps.disabled = disabled || formattedSecondProps.disabled;
    
    return (
        <div key={index} className={classnames("fields-wrapper", {horizontal, "with-background": withBackground})}>
            <div className="input-wrapper">
                <FirstFieldComponent name={`${name}.${index}.${firstKey}`} {...formattedFirstProps} />
                <SecondFieldComponent name={`${name}.${index}.${secondKey}`} {...formattedSecondProps} />
            </div>
            <div className={classnames("actions-wrapper", `actions-${index}`, {"with-labels": formattedFirstProps.label})}>
                <RoundIconContainer
                    name="add"
                    small
                    onClick={() => push({[firstKey]: firstEmptyValue, [secondKey]: secondEmptyValue})}
                    disabled={disabled}
                />
                <RoundIconContainer name="minus" small onClick={() => remove(index)} disabled={disabled || !allowRemove} />
            </div>
        </div>
    );
}

const ArrayField = (props) => {
    const {label, className, tooltipText} = props;
    const [field] = useField(props);
    
    const {name, value} = field;

    const {validateForm} = useFormikContext();

    useEffect(() => {
        validateForm();
    }, [value, validateForm])

    return (
        <div className={classnames("ps-field-wrapper", "ps-array-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <FieldArray name={name}>
                {({remove, push}) => value.map((item, index) => {
                    return (
                        <FieldItemWrapper key={index} {...props} name={name} index={index} value={value} remove={remove} push={push} />
                    )
                })}
            </FieldArray>
        </div>
    )
}

export default ArrayField;