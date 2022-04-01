// Import your modules here
// import demoModule from './demoModule';

import MODULE_TYPES from './MODULE_TYPES';

// Add your module to the modules array: const modules = [module1, module2, ...];
const modules = [
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
