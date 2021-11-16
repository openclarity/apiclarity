import React from 'react';
import Filter, { OPERATORS, METHOD_ITEMS, formatFiltersToQueryParams } from 'components/Filter';

export {
    formatFiltersToQueryParams
}

export const FILTERS_MAP = {
    method: {value: "method", label: "Method", valuesMapItems: METHOD_ITEMS, operators: [
        {...OPERATORS.is, valueItems: METHOD_ITEMS, creatable: false}
    ]},
    path: {value: "path", label: "Path", operators: [
        {...OPERATORS.contains, valueItems: [], creatable: true}
    ]},
}

const GeneralFilter = props => (<Filter {...props} filtersMap={FILTERS_MAP} />);

export default GeneralFilter;