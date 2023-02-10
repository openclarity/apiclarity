import React from 'react';
import { MODULE_TYPES } from '../MODULE_TYPES.js';
import { useFetch } from 'hooks';
import Loader from 'components/Loader';

const EventModule = props => {
    const {eventId} = props;
    const [{loading, data}] = useFetch(`modules/lineage/event/${eventId}/annotation/SAMPLE`);

    if (loading) {
        return <Loader />;
    }

    
    return (
        <div>Event Annotaton: <strong>{data}</strong></div>
    )
    
};

const APIModule = props => {
    const {inventoryId} = props;
    const [{loading, data}] = useFetch(`modules/lineage/api/${inventoryId}/annotation/SAMPLE`);

    if (loading) {
        return <Loader />;
    }
    
    return (
        <div>API Annotaton: <strong>{data}</strong></div>
    )
   
};

const lineageEvent = {
    name: 'lineage',
    component: EventModule,
    endpoint: '/lineage',
    type: MODULE_TYPES.EVENT_DETAILS
};

const lineageAPI = {
    name: 'lineage',
    component: APIModule,
    endpoint: '/lineage',
    type: MODULE_TYPES.INVENTORY_DETAILS
};

export {
    lineageEvent,
    lineageAPI,
}

// Make sure file ../index.js adds the following:
// import { lineageEvent, lineageAPI } from './lineage';
