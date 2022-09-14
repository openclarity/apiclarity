import React from 'react';

import COLORS from 'utils/scss_variables.module.scss';

import './dot-loader.scss';

const Dot = ({color}) => <div className="loader-dot" style={{backgroundColor: color}}></div>;

const DotLoader = ({color=COLORS["color-main"]}) => (
    <div className="dot-loader-wrapper">
        <Dot color={color} />
        <Dot color={color} />
        <Dot color={color} />
    </div>
);

export default DotLoader;
