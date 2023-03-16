import React from 'react';
import { Route, Switch, BrowserRouter, NavLink } from 'react-router-dom';
import Icon, { ICON_NAMES } from 'components/Icon';
import IconTemplates from 'components/Icon/IconTemplates';
import Notification from 'components/Notification';
import Dashboard from 'layout/Dashboard';
import Inventory from 'layout/Inventory';
import TraceSources from 'layout/TraceSources';
import Events from 'layout/Events';
import Reviewer from 'layout/Reviewer';
import { NotificationProvider, useNotificationState, useNotificationDispatch, removeNotification } from 'context/NotificationProvider'; 
import brandImage from 'utils/images/brand.svg';
import Paths from '../../Path'

import './app.scss';

const {
    ROOT_PATH,
    INVENTORY_ROOT_PATH,
    EVENTS_ROOT_PATH,
    TRACE_SOURCES_ROOT_PATH,
    REVIEWER_ROOT_PATH
} = Paths

const ROUTES = [
	{
		path: ROOT_PATH,
        exact: true,
		component: Dashboard,
        icon: ICON_NAMES.DASHBOARD
	},
	{
		path: INVENTORY_ROOT_PATH,
		component: Inventory,
        icon: ICON_NAMES.INVENTORY
	},
    {
		path: TRACE_SOURCES_ROOT_PATH,
		component: TraceSources,
        icon: ICON_NAMES.TRACE_SOURCE
	},
	{
		path: EVENTS_ROOT_PATH,
		component: Events,
        icon: ICON_NAMES.EVENTS
	},
    {
		path: REVIEWER_ROOT_PATH,
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
