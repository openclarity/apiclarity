import React from 'react';
import Select from 'react-select';
import CreatableSelect from 'react-select/creatable';
import classnames from 'classnames';

import COLORS from 'utils/scss_variables.module.scss';
import './dropdown-select.scss';

const DropdownSelect = ({items, value, onChange, creatable=false, disabled=false, placeholder="Select...", isMulti=false, className}) => {
    const SelectComponent = creatable ? CreatableSelect : Select;
    
    return (
        <SelectComponent
            value={value}
            onChange={onChange}
            className={classnames("dropdown-select", className)}
            classNamePrefix="dropdown-select"
            options={items}
            isClearable={false}
            isDisabled={disabled}
            placeholder={placeholder}
            isMulti={isMulti}
            styles={{
                control: (provided) => ({
                    ...provided,
                    // height: 36,
                    minHeight: 36,
                    borderRadius: 2,
                    borderColor: COLORS["color-grey-light"],
                    boxShadow: "none",
                    "&:hover": {
                        ...provided["&:hover"],
                        borderColor: COLORS["color-grey-light"]
                    },
                    backgroundColor: "white",
                    cursor: "pointer",
                    fontSize: "14px",
                    lineHeight: "18px"
                }),
                option: (provided, state) => {
                    const {isSelected, isDisabled} = state;
                    
                    return ({
                        ...provided,
                        color: isSelected ? COLORS["color-grey-dark"] : (isDisabled ? COLORS["color-grey-light"] : COLORS["color-grey-dark"]),
                        backgroundColor: isSelected ? COLORS["color-grey-lighter"] : "transparent",
                        fontWeight: isSelected ? "bold" : "normal",
                        cursor: "pointer"
                    });
                },
                placeholder: (provided, state) => ({
                    ...provided,
                    color: state.isDisabled ? COLORS["color-grey"] : COLORS["color-grey-dark"]
                }),
                menu: (provided) => ({
                    ...provided,
                    borderRadius: 2,
                    border: `1px solid ${COLORS["color-grey"]}`,
                    borderTop: `2px solid ${COLORS["color-main-light"]}`,
                    fontSize: "14px",
                    lineHeight: "18px"
                }),
                multiValueLabel: (provided) => ({
                    ...provided,
                    color: COLORS["color-grey-black"],
                    backgroundColor: COLORS["color-main-light"],
                    borderRadius: 0
                }),
                multiValueRemove: (provided) => ({
                    ...provided,
                    ":hover": {
                        ...provided[":hover"],
                        color: COLORS["color-grey-black"],
                        backgroundColor: COLORS["color-main-light"],
                        cursor: "pointer"
                    },
                    color: COLORS["color-grey-black"],
                    backgroundColor: COLORS["color-main-light"],
                    borderRadius: 0
                })
            }}
        />
    )
}

export default DropdownSelect;