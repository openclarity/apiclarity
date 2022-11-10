import React, { useState } from 'react';
import classnames from 'classnames';
import moment from 'moment';
import { isNull } from 'lodash';
import { OPEN_DOWN, OPEN_UP } from 'react-dates/constants';
import { SingleDatePicker } from 'react-dates';
import { isEmpty } from 'lodash';
import { useField } from 'formik';
import Arrow from 'components/Arrow';
import { FieldLabel, FieldError } from '../utils';

import 'react-dates/lib/css/_datepicker.css';
import './date-field.scss';

const DATE_FORMAT = 'YYYY-MM-DD';

const DateField = (props) => {
    const {label, className, tooltipText, displayFormat="MMM Do", isFullWidth=false, placeholder="Date", openDown=false} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    const [focused, setFocused] = useState(false);

    const formattedValue = !!value ? moment(value) : null;

    return (
        <div className={classnames("ps-field-wrapper", "ps-date-field-wrapper", {[className]: className}, {"full-width": isFullWidth})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <div className="selector-wrapper">
                <SingleDatePicker
                    date={formattedValue}
                    onDateChange={date => setValue(isNull(date) ? "" : moment(date).format(DATE_FORMAT))}
                    focused={focused}
                    onFocusChange={({focused}) => setFocused(focused)}
                    id={name}
                    daySize={30}
                    numberOfMonths={1}
                    hideKeyboardShortcutsPanel={true}
                    openDirection={openDown ? OPEN_DOWN : OPEN_UP}
                    small={true}
                    displayFormat={displayFormat}
                    navPrev={<Arrow name="left" />}
                    navNext={<Arrow name="right" />}
                    placeholder={placeholder}
                    readOnly={true}
                />
            </div>
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default DateField;