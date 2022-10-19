import { MODULE_TYPES } from '../MODULE_TYPES.js';
import FuzzingTab from './FuzzingTab';

const pluginAPIDetails = {
    name: 'Testing',
    component: FuzzingTab,
    endpoint: '/fuzzer',
    type: MODULE_TYPES.INVENTORY_DETAILS
};

export {
    pluginAPIDetails
};
