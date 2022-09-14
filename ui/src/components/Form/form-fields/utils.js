import React from 'react';
import classnames from 'classnames';
import InfoIcon from 'components/InfoIcon';

export const FieldLabel = ({children, tooltipId, tooltipText, className}) => (
    <div className={classnames("ps-field-label-wrapper", {[className]: !!className})}>
        <label htmlFor={tooltipId} className="form-label">{children}</label>
        {!!tooltipText && <InfoIcon tooltipId={tooltipId} text={tooltipText} />}
    </div>
);

export const FieldDescription = ({children}) => (
    <div className="ps-field-description">{children}</div>
);

export const FieldError = ({children}) => (
    <div className="ps-field-error">{children}</div>
)