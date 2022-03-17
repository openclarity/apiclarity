import React from 'react';
import PLUGIN_TYPES from '../PLUGIN_TYPES.js';

const DemoPlugin = () => {
    return (
        <div>Demo Plugin</div>
    );
};

export default {
    name: 'Demo Plugin',
    component: DemoPlugin,
    endpoint: '/demoplugin',
    type: PLUGIN_TYPES.EVENT_DETAILS
};
