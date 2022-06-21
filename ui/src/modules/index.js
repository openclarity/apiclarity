import { MODULE_TYPES, MODULE_STATUS_TYPES_MAP } from './MODULE_TYPES';

import { bfla, bflaApiInventory } from './bfla';
import { pluginEventDetails as taEventDetails, pluginAPIDetails as taAPIDetails } from './traceanalyzer';
import { pluginAPIDetails as fuzzerAPIDetails } from './fuzzer';


// Add your module to the modules array: const modules = [module1, module2, ...];
const modules = [
    taEventDetails, taAPIDetails,
    fuzzerAPIDetails,
    bfla, bflaApiInventory
];


// utility for core components to find the modules based on their type.
const getModules = (type) => {
    return modules.reduce((accum, m) => {
        if (m.type === type) {
            accum.push(m);
        }

        return accum;
    }, []);
};

export { getModules, MODULE_TYPES, MODULE_STATUS_TYPES_MAP };
export default modules;
