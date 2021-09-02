import React from 'react';
import { useHistory } from 'react-router-dom';
import Arrow, { ARROW_NAMES } from 'components/Arrow';

import './back-route-button.scss';

const BackRouteButton = ({title, path, query}) => {
    const history = useHistory();

    return (
        <div className="back-route-button" onClick={() => history.push({pathname: path, query})}>
            <Arrow name={ARROW_NAMES.LEFT} />
            <div className="back-title">{title}</div>
        </div>
    )
}

export default BackRouteButton;