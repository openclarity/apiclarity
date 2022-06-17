import React from 'react';
import Filter, { OPERATORS, METHOD_ITEMS, ALERT_ITEMS, formatFiltersToQueryParams } from 'components/Filter';
import { SPEC_DIFF_TYPES_MAP } from 'components/SpecDiffIcon';
import { getModules, MODULE_TYPES } from 'modules';

export {
    formatFiltersToQueryParams
}

const MODULE_ALERT_FILTERS = getModules(MODULE_TYPES.EVENT_DETAILS).map((m) => ({ value: m.moduleName, label: m.name }));

const SPEC_DIFF_ITEMS = [
    {value: "true", label: "present"},
    {value: "false", label: "not present"},
];

const FILTERS_MAP = {
    method: {value: "method", label: "Method", valuesMapItems: METHOD_ITEMS, operators: [
        {...OPERATORS.is, valueItems: METHOD_ITEMS, creatable: false}
    ]},
    path: {value: "path", label: "Path", operators: [
        {...OPERATORS.is, valueItems: [], creatable: true},
        {...OPERATORS.isNot, valueItems: [], creatable: true},
        {...OPERATORS.start},
        {...OPERATORS.end},
        {...OPERATORS.contains, valueItems: [], creatable: true}
    ]},
    statusCode: {value: "statusCode", label: "Status code", operators: [
        {...OPERATORS.is, valueItems: [], creatable: true},
        {...OPERATORS.isNot, valueItems: [], creatable: true},
        {...OPERATORS.gte},
        {...OPERATORS.lte},
    ]},
    sourceIP: {value: "sourceIP", label: "Source", operators: [
        {...OPERATORS.is, valueItems: [], creatable: true},
        {...OPERATORS.isNot, valueItems: [], creatable: true}
    ]},
    destinationIP: {value: "destinationIP", label: "Destination", operators: [
        {...OPERATORS.is, valueItems: [], creatable: true},
        {...OPERATORS.isNot, valueItems: [], creatable: true}
    ]},
    destinationPort: {value: "destinationPort", label: "Destination port", operators: [
        {...OPERATORS.is, valueItems: [], creatable: true},
        {...OPERATORS.isNot, valueItems: [], creatable: true}
    ]},
    spec: {value: "spec", label: "Host", operators: [
        {...OPERATORS.is, valueItems: [], creatable: true},
        {...OPERATORS.isNot, valueItems: [], creatable: true},
        {...OPERATORS.start},
        {...OPERATORS.end},
        {...OPERATORS.contains, valueItems: [], creatable: true}
    ]},
    hasSpecDiff: {value: "hasSpecDiff", label: "Spec diff", valuesMapItems: SPEC_DIFF_ITEMS, operators: [
        {...OPERATORS.is, valueItems: SPEC_DIFF_ITEMS, creatable: false, isSingleSelect: true},
    ]},
    specDiffType: {value: "specDiffType", label: "Spec diff type", operators: [
        {...OPERATORS.is, valueItems: Object.values(SPEC_DIFF_TYPES_MAP), creatable: false}
    ]},
    alertType: {value: "alertType", label:  "Alert Type", operators: [
        {...OPERATORS.is, valueItems: MODULE_ALERT_FILTERS, creatable: false}
    ]},
    alert: {value: "alert", label: "Alert Level", valuesMapItems: ALERT_ITEMS, operators: [
        {...OPERATORS.is, valueItems: ALERT_ITEMS, creatable: false}
    ]},
}

const GeneralFilter = props => (<Filter {...props} filtersMap={FILTERS_MAP} />);

export default GeneralFilter;
