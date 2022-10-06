import React, { useState } from 'react';
import classnames from 'classnames';
import { Route, Switch, Redirect, useLocation, useRouteMatch, NavLink } from 'react-router-dom';
import { useFetch } from 'hooks';

import Loader from 'components/Loader';
import FindingsTable from './FindingsTable';
import TestsTable from './TestsTable';
import TestOverview from './TestOverview';
import { MODULE_TYPES } from '../MODULE_TYPES.js';
import FuzzingTab from './FuzzingTab';

import './tests.scss';

export const TESTING_TAB_ITEMS = {
    FINDINGS: {
        value: "FINDINGS",
        label: "Findings",
        dataKey: "providedSpec",
        exact: true,
        component: FindingsTable,
        urlSuffix: "/findings",
        resetConfirmationText: "Resetting the provided spec will result in loss of the uploaded spec."
    },
    TESTS: {
        value: "TESTS",
        exact: true,
        label: "Tests",
        dataKey: "reconstructedSpec",
        component: TestsTable,
        urlSuffix: "/tests",
        resetConfirmationText: "Resetting the reconstructed spec will result in loss of spec and the information that was used to reconstruct it. If reset, to reconstruct again generate the relevant API traffic and review."
    }
}

const InnerTabs = ({selected, items, onSelect, url}) => {
    return (
        <div className="spec-inner-tabs-wrapper">
            {
                items.map(({ value, label, urlSuffix }, index) => (
                    <NavLink key={index} className={classnames("inner-tab-item")} to={`${url}${urlSuffix}`}>{label}</NavLink>
                ))
            }
        </div>
    );
 };

const FuzzerAPIDetails = props => {
    const {url, path} = useRouteMatch();
    const {query} = useLocation();
    const {inititalSelectedTab=TESTING_TAB_ITEMS.TESTS.value} = query || {};
    const [selectedTab, setSelectedTab] = useState(inititalSelectedTab);

    var inventoryId = props.inventoryId;

    // Get inventoryName
    const [{loading, data}] = useFetch("apiInventory", {queryParams: {apiId: inventoryId, type: "INTERNAL", page: 1, pageSize: 1, sortKey: "name"}});
    if (loading) {
        return <Loader />;
    }
    if (!data.items) {
        return null;
    }

    const inventoryName = data.items[0].name;

    return (
        <div className="inventory-details-spec-wrapper">
            <React.Fragment>
                <InnerTabs selected={selectedTab} items={Object.values(TESTING_TAB_ITEMS)} url={url} onSelect={selected => setSelectedTab(selected)} />
                <Switch>
                    <Redirect exact from={url} to={`${url}/tests`} />
                    <Route path={`${path}/findings`} component={() => <FindingsTable inventoryId={inventoryId} inventoryName={inventoryName} />} />
                    <Route path={`${path}/tests/:startTime`} component={() => <TestOverview backUrl={url} />} />
                    <Route path={`${path}/tests`} exact component={() => <TestsTable inventoryId={inventoryId} inventoryName={inventoryName}/>} />
                </Switch>
            </React.Fragment>
        </div>

    );
};

const pluginAPIDetails = {
    name: 'Testing',
    component: FuzzingTab,
    // component: FuzzerAPIDetails,
    endpoint: '/fuzzer',
    type: MODULE_TYPES.INVENTORY_DETAILS
}

export {
    pluginAPIDetails
}
