import React from 'react';
import classnames from 'classnames';
import { useField } from 'formik';
import ToggleButton from 'components/ToggleButton';
import InfoIcon from 'components/InfoIcon';

import './toggle-field.scss';

const ToggleField = (props) => {
    const {label, className, disabled, tooltipText, width} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    return (
        <div className={classnames("ps-field-wrapper", "ps-toogle-field-wrapper", {[className]: className})}>
            <ToggleButton
                {...meta}
                title={label}
                onChange={setValue}
                checked={value}
                width={width}
                disabled={disabled}
            />
            {!!tooltipText && <InfoIcon tooltipId={name} text={tooltipText} />}
        </div>
    )
}

export default ToggleField;