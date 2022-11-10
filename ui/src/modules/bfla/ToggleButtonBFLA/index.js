import React from 'react';
import Toggle from 'react-toggle';
import { isEmpty } from 'lodash';
import classnames from 'classnames';

import 'react-toggle/style.css';
import './toggle-button-bfla.scss';

const Title = ({ title, withLeftMargin, withBoldTitle }) => {
    return !isEmpty(title) && <span className={classnames("toggle-title", { "with-left-margin": withLeftMargin }, { "with-bold-text": withBoldTitle })}>{title}</span>;
};

const ToggleButton = (props) => {
    const { title, withBoldTitle, checked = false, onChange, disabled, fullWidth = false, width, className, secondaryTitle, small } = props;

    const labelProps = {
        className: classnames(
            "toggle-container",
            { "full-width": fullWidth },
            { [className]: !!className },
            { "with-secondary": !!secondaryTitle },
            { disabled: !!disabled },
            { small }
        )
    };

    if (!!width) {
        labelProps.style = { width };
    }

    return (
        <label {...labelProps}>
            {!!secondaryTitle ? <Title title={secondaryTitle} withBoldTitle={withBoldTitle} /> : <Title title={title} withBoldTitle={withBoldTitle} />}
            <Toggle
                icons={false}
                checked={checked}
                onChange={({ target }) => onChange(target.checked)}
                disabled={disabled}
            />
            {!!secondaryTitle && <Title title={title} withLeftMargin />}
        </label>
    );
}

export default ToggleButton;
