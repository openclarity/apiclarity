import React from 'react';
import { Route, Switch, BrowserRouter, NavLink } from 'react-router-dom';
import Icon, { ICON_NAMES } from 'components/Icon';
import IconTemplates from 'components/Icon/IconTemplates';
import Notification from 'components/Notification';
import Dashboard from 'layout/Dashboard';
import Inventory from 'layout/Inventory';
import Events from 'layout/Events';
import Reviewer from 'layout/Reviewer';
import { NotificationProvider, useNotificationState, useNotificationDispatch, removeNotification } from 'context/NotificationProvider'; 
import brandImage from 'utils/images/brand.svg';

import './app.scss';

const ROUTES = [
	{
		path: "/",
        exact: true,
		component: Dashboard,
        icon: ICON_NAMES.DASHBOARD
	},
	{
		path: "/inventory",
		component: Inventory,
        icon: ICON_NAMES.INVENTORY
	},
	{
		path: "/events",
		component: Events,
        icon: ICON_NAMES.EVENTS
	},
    {
		path: "/reviewer",
		component: Reviewer,
        noLink: true
	}
];

const ConnectedNotification = () => {
    const {message, type} = useNotificationState();
    const dispatch = useNotificationDispatch()

    if (!message) {
        return null;
    }

    return <Notification message={message} type={type} onClose={() => removeNotification(dispatch)} />
}

const App = () => (
    <div className="app-wrapper">
        <div className="top-bar-container">
            <img src={brandImage} alt="APIClarity" /> 
        </div>
        <NotificationProvider>
            <div id="main-wrapper">
                <BrowserRouter>
                    <div className="sidebar-container">
                        {
                            ROUTES.filter(route => !route.noLink).map(({path, icon, exact}) => (
                                <NavLink className="nav-item" key={path} to={path} exact={exact}>
                                    <Icon name={icon} />
                                </NavLink>
                            ))
                        }
                    </div>
                    <main role="main">
                        <Switch>
                            {
                                ROUTES.map(({path, component, exact}) => (
                                    <Route key={path} path={path} exact={exact} component={component} />
                                ))
                            }
                        </Switch>
                    </main>
                </BrowserRouter>
            </div>
            <ConnectedNotification />
        </NotificationProvider>
        <IconTemplates />
    </div>
)

export default App;
