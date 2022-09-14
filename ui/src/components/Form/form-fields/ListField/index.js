import React from 'react';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import { FieldArray, useField } from 'formik';
import RoundIconContainer from 'components/RoundIconContainer';
import { FieldLabel } from '../utils';

import './list-field.scss';

const ListField = (props) => {
    const {label, fieldComponent: FieldComponent, fieldProps, className, tooltipText, disabled} = props;
    const [field] = useField(props);
    
    const {name, value} = field;

    const {customComponenet: CustomComponenet, ...otherProps} = fieldProps;

    const allowRemove = value.length > 1;

    return (
        <div className={classnames("ps-field-wrapper", "ps-list-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <FieldArray name={name}>
                {({remove, push}) => value.map((item, index) => (
                    <div key={index} className="field-wrapper">
                        <div className="input-wrapper">
                            <FieldComponent name={`${name}.${index}`} {...otherProps} disabled={disabled} />
                            {!!CustomComponenet && <CustomComponenet value={item} /> }
                        </div>
                        <div className={classnames("actions-wrapper", `actions-${index}`)}>
                            <RoundIconContainer name="add" small onClick={() => push("")} disabled={disabled} />
                            <RoundIconContainer name="minus" small onClick={() => remove(index)} disabled={disabled || !allowRemove} />
                        </div>
                    </div>
                ))}
            </FieldArray>
        </div>
    )
}

export default ListField;