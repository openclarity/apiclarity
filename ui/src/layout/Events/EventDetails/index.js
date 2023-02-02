import React from 'react';
import { useParams, useRouteMatch } from 'react-router-dom';
import BackRouteButton from 'components/BackRouteButton';
import Title from 'components/Title';
import TabbedPageContainer from 'components/TabbedPageContainer';
import PageContainer from 'components/PageContainer';
import Loader from 'components/Loader';
import RoundIconContainer from 'components/RoundIconContainer';
import { useFetch } from 'hooks';
import { formatDate } from 'utils/utils';
import { getModules, MODULE_TYPES, MODULE_STATUS_TYPES_MAP } from 'modules';
import Details from './Details';
import Specs from './Specs';

import Icon from 'components/Icon';

import './event-details.scss';

const createTabAlertTitle = (id, title, status) => (
    <div className="module-tab-title-wrapper">
        <div className="module-tab-title">{title}</div>
        <span>{status ? ModuleStatusIcon(id, status.reason) : ""}</span>
    </div>
);

const EventDetails = () => {
    const {path, url, isExact} = useRouteMatch();
    const params = useParams();
    const {eventId} = params;

    const [{loading, data}] = useFetch(`apiEvents/${eventId}`);

    if (loading) {
        return <Loader />;
    }

    if (!data) {
        return null;
    }

    const {time, alerts=[]} = data;
    const modules = getModules(MODULE_TYPES.EVENT_DETAILS);
    const moduleTabs = modules.map((m, idx) => {
        const MStatus = alerts.find((a) => a.moduleName === m.moduleName);
        return {
            title: createTabAlertTitle(idx, m.name, MStatus),
            linkTo: `${url}${m.endpoint}`,
            to: `${path}${m.endpoint}`,
            component: () => <m.component  {...{...data, eventId}}/>
        };
    });

    return (
        <div className="events-details-page">
            <BackRouteButton title="API events" path={url.replace(`/${eventId}`, "")} />
            <Title>{formatDate(time)}</Title>
            <div className="tabbed-container-wrapper">
                <PageContainer className="fixed-sidebar">
                    <div className="sidebar-heading">
                        <div className="title"><div>Event Summary</div></div>
                    </div>
                    <div className="sidebar-content">
                        <Details data={data} />
                    </div>
                </PageContainer>
                <div className="event-details-tab-container">
                    <TabbedPageContainer
                        items={[
                            { title: "Spec", linkTo: url, to: path, exact: true, component: () => <Specs data={data} /> },
                            ...moduleTabs
                        ]}
                        noContentMargin={isExact}
                    />
                </div>
            </div>
        </div>
    )
}

const ModuleStatusIcon = (id, statusType) => {
    const tooltipId = `module-status-${id}`;
    // const {icon, tooltip, color} = MODULE_STATUS_TYPES_MAP[statusType] || MODULE_STATUS_TYPES_MAP['NO_STATUS'];
    const {icon, color, value} = MODULE_STATUS_TYPES_MAP[statusType] || {};
    const isInfo = value === MODULE_STATUS_TYPES_MAP.ALERT_INFO.value;

    return (
        <div className="module-status-icon" style={{ width: "22px" }}>
                <React.Fragment>
                    <div data-tip data-for={tooltipId}>{isInfo ? <RoundIconContainer name={icon} /> : <Icon name={icon} style={{ color }} />}</div>
                </React.Fragment>
        </div>
    );
};

export default EventDetails;
