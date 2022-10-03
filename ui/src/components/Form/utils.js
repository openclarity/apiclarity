import React, { useState } from 'react';
import classnames from 'classnames';
import RoundIconContainer from 'components/RoundIconContainer';
import Arrow from 'components/Arrow';
import Icon, { ICON_NAMES } from 'components/Icon';

export const FieldAlert = ({message}) => (
    <div className="form-field-alert-wrapper">
        <RoundIconContainer name={ICON_NAMES.EXCLAMATION_MARK} />
        <div className="alert-text">{message}</div>
    </div>
);

export const InnerFieldWrapper = ({title, isVisable, children, isSecondLayer=false}) => {
    const [isOpen, setIsOpen] = useState(true);

    if (!isVisable) {
        return null;
    }

    return (
        <div className={classnames("form-inner-field-wrapper", {"is-second-layer": isSecondLayer})}>
            <div className="form-inner-field-header" onClick={() => setIsOpen(!isOpen)}>
                <div className="header-title">{title}</div>
                <Arrow name={isOpen ? "top" : "bottom"} small={true} />
            </div>
            {isOpen && <div className="form-inner-field-content">{children}</div>}
        </div>
    )
}

export const FormNotificationMessage = ({children, className, secondary=false, isError=false}) => (
    <div className={classnames("form-message-container", className, {secondary}, {error: isError})}>
        {/* <Icon name={secondary ? ICON_NAMES.ALERT : ICON_NAMES.EXCLAMATION_MARK_ROUND} /> */}
        <Icon name={secondary ? ICON_NAMES.ALERT : ICON_NAMES.INFO} />
        <div className="message">{children}</div>
    </div>
);

export const FormInstructionsTitle = ({children}) => (
    <div className="form-insturctions-title-container">
        <div className="title-bullet"></div>
        <div className="title">{children}</div>
    </div>
);
