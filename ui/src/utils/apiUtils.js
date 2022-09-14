import { isEmpty } from 'lodash';

export const queryString = (params) => Object.keys(params).map((key) => {
    return encodeURIComponent(key) + '=' + encodeURIComponent(params[key])
}).join('&');

export const valueToValueLabel = (items) => items.sort().map(item => ({value: item, label: item}));
export const valueToValueLabelFromProp = (items, propName) => valueToValueLabel([...new Set(items.map(item => item[propName]))]);
export const nameIdToValueLabel = (items) => items.map(({id, name}) => ({value: id, label: name}));

export const AUTHENTICATION_ERROR_CODES = [401, 403];
const AUTHENTICATION_ERROR_INDICATOR = "AUTHENTICATION_ERROR_INDICATOR";
export const dataFetcher = ({url, queryParams, successCallback, errorCallback, authenticationErrorCallback}) => {
    const fullUrl = isEmpty(queryParams) ? `/api/${url}` : `/api/${url}?${queryString(queryParams)}`;

    fetch(fullUrl, {credentials: 'include'})
        .then(response => {
            if (AUTHENTICATION_ERROR_CODES.includes(response.status) && !!authenticationErrorCallback) {
                authenticationErrorCallback();
                throw Error(AUTHENTICATION_ERROR_INDICATOR);
            }

            if (!response.ok) {
                throw Error(response.statusText);
            }

            return response;
        })
        .then(response => response.json())
        .then(data => {
            if (!!successCallback) {
                successCallback(data);
            }
        })
        .catch(error => {
            if (!!error && error.message === AUTHENTICATION_ERROR_INDICATOR) {
                //no error display on authentication error
                return;
            }

            if (!!errorCallback) {
                errorCallback(error);
            }
        });
}
export async function asyncDataFetcher(props) {
    return dataFetcher(props);
}

export async function dataMultiFetcher({urlsData, successCallback, errorCallback, authenticationErrorCallback}) {
    try {
        const response = await Promise.all(
            urlsData.map(urlData => {
                const {url, queryParams, data, method="GET"} = urlData;

                const fullUrl = isEmpty(queryParams) ? `/api/${url}` : `/api/${url}?${queryString(queryParams)}`;
                const formattedMethod = method.toUpperCase()
                const options = {
                    credentials: 'include',
                    method: formattedMethod
                };

                if (formattedMethod === "POST" || formattedMethod === "PUT") {
                    options.headers = {'content-type': 'application/json'};
                    options.body = JSON.stringify(data);
                }

                return fetch(fullUrl, options)
                    .then(response => {
                        if (AUTHENTICATION_ERROR_CODES.includes(response.status) && !!authenticationErrorCallback) {
                            authenticationErrorCallback();
                            throw Error(AUTHENTICATION_ERROR_INDICATOR);
                        }

                        if (!response.ok) {
                            throw Error(response.statusText);
                        }

                        return response;
                    })
                    .then(response => response.json())
            })
        );

        const data = response.map((item, index) => ({key: urlsData[index].key, data: item})).reduce((accumulator, curr) => {
            accumulator[curr.key] = curr.data;
            return accumulator
        }, {});

        if (!!successCallback) {
            successCallback(data);
        }
    } catch (error) {
        if (!!error && error.message === AUTHENTICATION_ERROR_INDICATOR) {
            //no error display on authentication error
            return;
        }

        if (!!errorCallback) {
            errorCallback(error);
        }
    }
}
