import React from 'react';
import { useParams, useRouteMatch } from 'react-router-dom';
import BackRouteButton from 'components/BackRouteButton';
import Title from 'components/Title';
import TabbedPageContainer from 'components/TabbedPageContainer';
import Loader from 'components/Loader';
import { useFetch } from 'hooks';
import Specs from './Specs';
import { getModules, MODULE_TYPES } from 'modules';

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

    const modules = getModules(MODULE_TYPES.INVENTORY_DETAILS);
    const moduleTabs = modules.map((m) => {
        return {
            title: m.name,
            linkTo: `${url}${m.endpoint}`,
            to: `${path}${m.endpoint}`,
            component: () => <m.component  {...{...data.items[0], inventoryId, type, outerHistory: url}}/>
        };
    });


    return (
        <div>
            <BackRouteButton title="API inventory" path={url.replace(`/${inventoryId}`, "")} />
            <Title>{inventoryName}</Title>
            <TabbedPageContainer
                items={[
                    {title: "Spec", linkTo: url, to: path, exact: true, component: () => <Specs inventoryId={inventoryId} inventoryName={inventoryName} />},
                    ...moduleTabs
                ]}
                noContentMargin={true}
            />
        </div>
    )
}

export default InventoryDetails;
