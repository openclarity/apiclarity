import React, { useEffect, useState } from 'react';
import { isUndefined, isEmpty } from 'lodash';
import { Formik, Form, useFormikContext } from 'formik';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';
import Button from 'components/Button';
import SelectField from './filter-fields/SelectField';
import MultiselectField from './filter-fields/MultiselectField';
import TextField from './filter-fields/TextField';
import { OPERATORS, METHOD_ITEMS, ALERT_ITEMS, formatFiltersToQueryParams, getValueLabel } from './utils';

import './filter.scss';

export {
    OPERATORS,
    METHOD_ITEMS,
    ALERT_ITEMS,
    formatFiltersToQueryParams
}

const FormFields = ({onAdd, filtersMap, existingFilters}) => {
    const {values: formValues, setFieldValue, resetForm} = useFormikContext();
    const {scope, operator, value} = formValues;

    const selectedScopeData = filtersMap[scope] || {};
    const inUseOperatorsByScope = existingFilters.filter(filterItem => filterItem.scope === scope).map(({operator}) => operator);
    const operatorByScopeItems = selectedScopeData.operators || [];
    const formattedOperatorByScopeItems = operatorByScopeItems.map(operatorItem =>
        ({...operatorItem, isDisabled: inUseOperatorsByScope.includes(operatorItem.value)}))
    const selectedOperatorData = operatorByScopeItems.find(item => item.value === operator);
    const {valueItems, creatable, isSingleSelect} = selectedOperatorData || {};
    const ValueField = isUndefined(valueItems) ? TextField : (isSingleSelect ? SelectField : MultiselectField);
    const valuePlaceholder = isUndefined(valueItems) ? "Enter value..." : "Select value...";

    useEffect(() => {
        setFieldValue("operator", "");
        setFieldValue("value", "");
    }, [scope, setFieldValue]);

    useEffect(() => {
        setFieldValue("value", "");
    }, [operator, setFieldValue]);

    return (
        <React.Fragment>
            <SelectField
                name="scope"
                items={Object.values(filtersMap)}
                placeholder="Select scope..."
            />
            <SelectField
                name="operator"
                items={formattedOperatorByScopeItems}
                placeholder="Select operator..."
                disabled={!scope}
            />
            <ValueField
                className="filter-field-value"
                name="value"
                items={valueItems}
                placeholder={valuePlaceholder}
                creatable={creatable}
                disabled={!operator}
            />
            <Button disabled={isEmpty(value)} className="add-filter-button" onClick={() => {
                onAdd(formValues);
                resetForm();
            }}>OK</Button>
        </React.Fragment>
    )
}

const GeneralFilter = ({filters, onFilterUpdate, filtersMap, customFilterDisplay: CustomFilterDisplay}) => {
    const [showFiltersForm, setShowFiltersForm] = useState(false);

    return (
        <div className="general-filter-wrapper">
            <Button className={classnames("show-filters-button", {pressed: showFiltersForm})} onClick={() => setShowFiltersForm(!showFiltersForm)}>
                <Icon name={ICON_NAMES.FILTER} />
                Filters
            </Button>
            {showFiltersForm &&
                <div className="filter-form-wrapper">
                    <Formik
                        initialValues={{
                            scope: "",
                            operator: "",
                            value: ""
                        }}
                    >
                        <Form className="filter-form">
                            <FormFields
                                onAdd={filterData => onFilterUpdate([...filters, filterData])}
                                filtersMap={filtersMap}
                                existingFilters={filters}
                            />
                        </Form>
                    </Formik>
                    {!!CustomFilterDisplay && <div className="custom-filter-display-wrapper"><CustomFilterDisplay /></div>}
                </div>
            }
            <div className={classnames("filters-display-wrapper", {"is-empty": isEmpty(filters)})}>
                {
                    filters.map(({scope, operator, value}, index) => {
                        const {label: scopeLabel, valuesMapItems} = filtersMap[scope];
                        const operatorLabel = OPERATORS[operator].label;
                        const formattedValue = Array.isArray(value) ?
                            value.map(valueItem => getValueLabel(valuesMapItems, valueItem)).join(" or ") : getValueLabel(valuesMapItems, value);
                            
                        return (
                            <div className="filter-display-item" key={index}>
                                <span>{`${scopeLabel} ${operatorLabel} `}<span style={{fontWeight: "bold"}}>{formattedValue}</span></span>
                                <Icon
                                    name={ICON_NAMES.X_MARK}
                                    onClick={() => {
                                        const newFilters = filters.filter(filterItem => !(filterItem.scope === scope && filterItem.operator === operator));
                                        
                                        onFilterUpdate(newFilters); 
                                    }}
                                />
                            </div>
                        )
                    })
                }
                {!isEmpty(filters) && <Button secondary onClick={() => onFilterUpdate([])} >Delete all filters</Button>}
            </div>
        </div>
    );
}

export default GeneralFilter;
