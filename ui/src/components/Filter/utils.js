import {mapValues, keyBy} from 'lodash';

export const OPERATORS = {
    is: {value: "is", label: "is"},
    isNot: {value: "isNot", label: "is not"},
    start: {value: "start", label: "starts with"},
    end: {value: "end", label: "ends with"},
    contains: {value: "contains", label: "contains"},
    gte: {value: "gte", label: "greater than or equal to"},
    lte: {value: "lte", label: "less than or equal to"}
}

export const METHOD_ITEMS = [
    {value: "GET", label: "GET"},
    {value: "HEAD", label: "HEAD"},
    {value: "POST", label: "POST"},
    {value: "PUT", label: "PUT"},
    {value: "DELETE", label: "DELETE"},
    {value: "CONNECT", label: "CONNECT"},
    {value: "OPTIONS", label: "OPTIONS"},
    {value: "TRACE", label: "TRACE"},
    {value: "PATCH", label: "PATCH"}
];

export const ALERT_ITEMS = [
	{value: "ALERT_INFO", label: "INFORMATION"},
	{value: "ALERT_WARN", label: "WARNING"},
	{value: "ALERT_CRITICAL", label: "CRITICAL"},
]

export const formatFiltersToQueryParams = (filters) => {
    const filtersList = filters.map(({scope, operator, value} )=> ({key: `${scope}[${operator}]`, value}));

    return mapValues(keyBy(filtersList, "key"), "value");
};

export const getValueLabel = (valueItems=[], value) => {
    const valueItem = valueItems.find(valueItem => valueItem.value === value);

    return !!valueItem ? valueItem.label : value;
};
