import React from 'react';
import Filter, { OPERATORS, formatFiltersToQueryParams } from 'components/Filter';

export {
    formatFiltersToQueryParams
}

const HAS_SPEC_ITEMS = [
    {value: "true", label: "present"},
    {value: "false", label: "not present"},
];

const FILTERS_MAP = {
    name: {value: "name", label: "API name", operators: [
        {...OPERATORS.is, valueItems: [], creatable: true},
        {...OPERATORS.isNot, valueItems: [], creatable: true},
        {...OPERATORS.start},
        {...OPERATORS.end},
        {...OPERATORS.contains, valueItems: [], creatable: true}
    ]},
    port: {value: "port", label: "Port", operators: [
        {...OPERATORS.is, valueItems: [], creatable: true},
        {...OPERATORS.isNot, valueItems: [], creatable: true}
    ]},
    hasProvidedSpec: {value: "hasProvidedSpec", label: "Provided spec", valuesMapItems: HAS_SPEC_ITEMS, operators: [
        {...OPERATORS.is, valueItems: HAS_SPEC_ITEMS, creatable: false, isSingleSelect: true},
    ]},
    hasReconstructedSpec: {value: "hasReconstructedSpec", label: "Reconstructed spec", valuesMapItems: HAS_SPEC_ITEMS, operators: [
        {...OPERATORS.is, valueItems: HAS_SPEC_ITEMS, creatable: false, isSingleSelect: true},
    ]}
}

const GeneralFilter = props => (<Filter {...props} filtersMap={FILTERS_MAP} />);

export default GeneralFilter;