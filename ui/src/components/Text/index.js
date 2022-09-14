import React from 'react';
import classnames from 'classnames';

import './text.scss';

export const TEXT_TYPES = {
    TITLE_LARGE: "TITLE_LARGE",
    TITLE_MEDIUM: "TITLE_MEDIUM",
    H1: "H1",
    H2: "H2",
    H3: "H3",
    H4: "H4",
    BODY: "BODY",
    CTAS: "CTAS",
    LINKS: "LINKS",
    TABLE_BODY: "TABLE_BODY",
    TABLE_HEADER: "TABLE_HEADER",
    LEGENDS: "LEGENDS",
    FORM: "FORM",
    WIDGET_COUNTER: "WIDGET_COUNTER"
}

const TEXT_TYPE_STYLES_MAP = {
    TITLE_LARGE: {fontSize: "34px", lineHeight: "45px", fontWeight: "bold", textTransform: "none"},
    TITLE_MEDIUM: {fontSize: "26px", lineHeight: "30px", fontWeight: "bold", textTransform: "none"},
    H1: {fontSize: "14px", lineHeight: "24px", fontWeight: "bold", textTransform: "uppercase"},
    H2: {fontSize: "16px", lineHeight: "19px", fontWeight: "bold", textTransform: "none"},
    H3: {fontSize: "14px", lineHeight: "16px", fontWeight: "bold", textTransform: "none"},
    H4: {fontSize: "12px", lineHeight: "14px", fontWeight: "bold", textTransform: "none"},
    BODY: {fontSize: "14px", lineHeight: "18px", fontWeight: "normal", textTransform: "none"},
    CTAS: {fontSize: "10px", lineHeight: "10px", fontWeight: "bold", textTransform: "uppercase"},
    LINKS: {fontSize: "10px", lineHeight: "10px", fontWeight: "bold", textTransform: "none"},
    TABLE_BODY: {fontSize: "11px", lineHeight: "16px", fontWeight: "normal", textTransform: "none"},
    TABLE_HEADER: {fontSize: "9px", lineHeight: "12px", fontWeight: "bold", textTransform: "uppercase"},
    LEGENDS: {fontSize: "9px", lineHeight: "14px", fontWeight: "normal", textTransform: "none"},
    FORM: {fontSize: "13px", lineHeight: "20px", fontWeight: "normal", textTransform: "none"},
    WIDGET_COUNTER: {fontSize: "28px", lineHeight: "30px", fontWeight: "normal", textTransform: "none"}
}

const Text = ({children, type=TEXT_TYPES.BODY, withTopMargin=false, withBottomMargin=false, className, onClick}) => {
    if (!Object.keys(TEXT_TYPES).includes(type)) {
        console.error(`Text type '${type}' does not exist`);
    }

    const textClassName = classnames(
        "scn-text-wrapper",
        {"with-top-margin": withTopMargin},
        {"with-bottom-margin": withBottomMargin},
        {"clickable": !!onClick},
        {[className]: !!className}
    );

    return (
        <span className={textClassName} style={TEXT_TYPE_STYLES_MAP[type]} onClick={onClick}>
            {children}
        </span>
    );
}

export default Text;
