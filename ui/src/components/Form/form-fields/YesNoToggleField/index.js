import React from 'react';
import classnames from 'classnames';
import RadioField from '../RadioField';

import './yes-no-toggle-field.scss';

const YesNoToggleField = ({className, ...props}) => {
    return (
        <RadioField
            {...props}
            className={classnames("ps-yes-no-toggle", className)}
            items={[
                {value: true, label: "Yes"},
                {value: false, label: "No"},
            ]}
            horizontal={true}
        />
    )
}

export default YesNoToggleField;