### UI Modules

Modules give developers the ability to add functionality to the API Clarity UI without having to understand the internals of the API Clarity codebase.

Modules can be added to the API Inventory details view and API Event details view.  You're plugin will show up as a tab.  See the demoModule for a simple example of this.

Modules can have their own dependencies.  By adding a `package.json` to your module and listing dependencies, the API Clarity build process will install your dependencies.

### Create a new module

1. Create a new directory under `ui/src/modules`.
2. Export your module. You'll need to export it as an object with the following properties:

```json
export default {
    name: 'Demo Module', // this name will be used as the tab name in API Clarity
    component: DemoModule, // your component
    endpoint: '/demomodule', // the endpoint will be used by the router
    type: MODULE_TYPES.EVENT_DETAILS // the type indicates where your module is to appear in the UI.
};
```
The Module types are defined within `ui/src/module/MODULE_TYPES.js`.  The types are `EVENT_DETAILS` and `INVENTORY_DETAILS`.  If `type` is not exported (or using one that isn't defined within `MODULE_TYPES`) with your module, it will not be displayed in the UI.

3. Import your module in `ui/src/modules/index.js`.  Then add your imported module to the modules array.
