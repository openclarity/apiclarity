import React from 'react';
import MODULE_TYPES from '../MODULE_TYPES.js';

const DemoModule = () => {
    return (
        <div>Demo Module</div>
    );
};

export default {
    name: 'Demo Module',
    component: DemoModule,
    endpoint: '/demomodule',
    type: MODULE_TYPES.EVENT_DETAILS
};
