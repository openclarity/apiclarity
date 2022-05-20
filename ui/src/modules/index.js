import MODULE_TYPES from './MODULE_TYPES';
import { bfla, bflaApiInventory } from './bfla';
import { pluginEventDetails as taEventDetails, pluginAPIDetails as taAPIDetails } from './traceanalyzer';

const modules = [
    bfla, bflaApiInventory,
    taEventDetails, taAPIDetails,
    // demoModule
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

export { getModules, MODULE_TYPES };
export default modules;
