import React from 'react';
import { useParams, useRouteMatch } from 'react-router-dom';
import BackRouteButton from 'components/BackRouteButton';
import Title from 'components/Title';
import TabbedPageContainer from 'components/TabbedPageContainer';
import Loader from 'components/Loader';
import { useFetch } from 'hooks';
import Specs from './Specs';

const InventoryDetails = ({type}) => {
    const {path, url} = useRouteMatch();
    const params = useParams();
    const {inventoryId} = params;

    const [{loading, data}] = useFetch("apiInventory", {queryParams: {apiId: inventoryId, type, page: 1, pageSize: 1, sortKey: "name"}});

    if (loading) {
        return <Loader />;
    }

    if (!data.items) {
        return null;
    }

    const inventoryName = data.items[0].name;

    return (
        <div>
            <BackRouteButton title="API inventory" path={url.replace(`/${inventoryId}`, "")} />
            <Title>{inventoryName}</Title>
            <TabbedPageContainer
                items={[
                    {title: "Spec", linkTo: url, to: path, exact: true, component: () => <Specs inventoryId={inventoryId} inventoryName={inventoryName} />}
                ]}
                noContentMargin={true}
            />
        </div>
    )
}

export default InventoryDetails;