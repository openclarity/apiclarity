import React from 'react';
import { useParams, useRouteMatch } from 'react-router-dom';
import BackRouteButton from 'components/BackRouteButton';
import Title from 'components/Title';
import TabbedPageContainer from 'components/TabbedPageContainer';
import Loader from 'components/Loader';
import { useFetch } from 'hooks';
import { formatDate } from 'utils/utils';
import Details from './Details';
import SpecDiff from './SpecDiff';

import './event-details.scss';

const EventDetails = () => {
    const {path, url} = useRouteMatch();
    const params = useParams();
    const {eventId} = params;

    const [{loading, data}] = useFetch(`apiEvents/${eventId}`);

    if (loading) {
        return <Loader />;
    }

    if (!data) {
        return null;
    }

    const {time, hasProvidedSpecDiff, hasReconstructedSpecDiff} = data;
    
    return (
        <div className="events-details-page">
            <BackRouteButton title="API events" path={url.replace(`/${eventId}`, "")} />
            <Title>{formatDate(time)}</Title>
            <TabbedPageContainer
                items={[
                    {title: "Event details", linkTo: url, to: path, exact: true, component: () => <Details data={data} />},
                    {
                        title: "Reconstructed Spec Diff",
                        linkTo: `${url}/reconstructedDiffs`,
                        to: `${path}/reconstructedDiffs`,
                        component: () => <SpecDiff url={`apiEvents/${eventId}/reconstructedSpecDiff`} />,
                        disabled: !hasReconstructedSpecDiff
                    },
                    {
                        title: "Provided Spec Diff",
                        linkTo: `${url}/providedDiffs`,
                        to: `${path}/providedDiffs`,
                        component: () => <SpecDiff url={`apiEvents/${eventId}/providedSpecDiff`} />,
                        disabled: !hasProvidedSpecDiff
                    }
                ]}
            />
        </div>
    )
}

export default EventDetails;