import React, { useState } from 'react'
import AsyncSelect from 'react-select/async';
import { components } from 'react-select';
import { dataMultiFetcher, asyncDataFetcher } from 'utils/apiUtils';
import Button from 'components/Button';

import './add-client-bfla-field.scss';

const SearchItem = ({ label }) => {
    return (
        <div className="search-field-item">
            <div className="item-title">{label}</div>
        </div>
    );
}

export default function AddClientBFLAField(props) {
    const { name, disabled, handleAddNewClientToModel, selectedData } = props;

    const [selected, setSelected] = useState();
    const url = 'podDefinitions';
    const PLACEHOLDER_ADD_NEW_CLIENT = 'Add new client...';

    const getQueryParams = searchValue => ({ name: searchValue, offset: 0, maxResults: 50 })

    const promiseOptions = inputValue => {
        const isSingleFetch = !!URL;
        const fetcher = isSingleFetch ? asyncDataFetcher : dataMultiFetcher;

        return new Promise(resolve => fetcher({
            url,
            queryParams: isSingleFetch ? getQueryParams(inputValue) : null,
            successCallback: (data) => {
                resolve(data.map(({ name }) => ({
                    value: name,
                    label: name,
                })))
            },
            errorCallback: (error) => {
                // TODO handle
                // showErrorToaster();

                resolve([]);
            },
            authenticationErrorCallback: () => {
                // showErrorToaster();

                resolve([]);
            }
        }))
    }

    return (
        <div className='select-workload-wrapper'>
            <div className='select-workload-label'>
                Add new client
            </div>
            <AsyncSelect
                value={selected}
                name={name}
                onChange={
                    selected => {
                        setSelected(selected)
                    }
                }
                loadOptions={promiseOptions}
                placeholder={PLACEHOLDER_ADD_NEW_CLIENT}
                formatGroupLabel={group => <div className="search-group-label">{group.label}</div>}
                getOptionLabel={option => <SearchItem {...option} />}
                isClearable={true}
                isDisabled={disabled}
                noOptionsMessage={({ inputValue }) => {
                    if (inputValue === "") {
                        return PLACEHOLDER_ADD_NEW_CLIENT;
                    }

                    return "No options";
                }}
                components={{
                    SingleValue: ({ data, ...props }) =>
                        <components.SingleValue {...props}>{data.label}</components.SingleValue>
                }}
            />
            {selected &&
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'end' }}>
                    <Button className='select-workload-add-client-button' onClick={() => {
                        handleAddNewClientToModel(selected.value, selectedData)
                        setSelected("")
                    }}>
                        Add client
                    </Button>
                </div>
            }
        </div>
    )
}
