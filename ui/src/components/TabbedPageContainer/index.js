import React from 'react';
import classnames from 'classnames';
import { Route, NavLink } from 'react-router-dom';
import PageContainer from 'components/PageContainer';

import './tabbed-page-container.scss';

const TabbedPageContainer = ({items, noContentMargin}) => (
    <PageContainer>
        <div className="tabs-container">
            {
                items.map(({title, to, linkTo, exact, disabled}, index) => {
                    const TabTag = disabled ? "div" : NavLink;

                    return (
                        <TabTag key={index} className={classnames("tab-item", {disabled})} to={linkTo || to} exact={exact}>
                            <div>{title}</div>
                        </TabTag>
                    )
                })
            }
        </div>
        <div className={classnames("tab-content", {"with-margin": !noContentMargin})}>
            {
                items.map(({title, to, exact, component}, index) => <Route key={index} path={to} exact={exact} component={component} />)
            }
        </div>
    </PageContainer>
);

export default TabbedPageContainer;