import PLUGIN_TYPES from './PLUGIN_TYPES';

// Import your plugins here
// import demoPlugin from './demoPlugin';

// Add your plugin to the plugins array: const plugins = [plugin1, plugin2, ...];
// const plugins = [demoPlugin];
const plugins = [
];


// utility for core components to find the plugins based on their type.
const getPlugins = (type) => {
    return plugins.reduce((accum, p) => {
        if (p.type === type) {
            accum.push(p);
        }

        return accum;
    }, []);
};

export { getPlugins, PLUGIN_TYPES };
export default plugins;
