import React from "react";
import Filter, {
  OPERATORS,
  formatFiltersToQueryParams,
} from "components/Filter";
import { TRACES_TYPES } from "utils/utils";

export { formatFiltersToQueryParams };

const TYPE_ITEMS = Object.values(TRACES_TYPES);

const FILTERS_MAP = {
  name: {
    value: "name",
    label: "API name",
    operators: [
      { ...OPERATORS.start },
      { ...OPERATORS.end },
      { ...OPERATORS.contains },
    ],
  },
  type: {
    value: "type",
    label: "Type",
    operators: [
      {
        ...OPERATORS.is,
        valueItems: TYPE_ITEMS,
        creatable: true,
      },
      {
        ...OPERATORS.isNot,
        valueItems: TYPE_ITEMS,
        creatable: true,
      },
    ],
  },
};

const GeneralFilter = (props) => <Filter {...props} filtersMap={FILTERS_MAP} />;

export default GeneralFilter;
