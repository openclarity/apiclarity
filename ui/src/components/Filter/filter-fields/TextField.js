import React from 'react';
import classnames from 'classnames';
import { useField } from 'formik';

const TextField = ({className, ...props}) => {
    const [field] = useField(props);

    return (
        <input {...field} {...props} className={classnames("filter-form-field", "text-field", className)} />
    )
}

export default TextField;