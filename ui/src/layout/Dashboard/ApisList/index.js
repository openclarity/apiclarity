import React, { useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import Loader from 'components/Loader';
import { useFetch } from 'hooks';

import './apis-list.scss';

const ApisList = ({url, subColumn: {title, dataDisplay: DataDisplay}, getLink, apiIdKey, refreshTimestamp}) => {
    const history = useHistory();

    const [{loading, data, error}, fetchData] = useFetch(url);

    useEffect(() => {
        fetchData();
    }, [fetchData, refreshTimestamp]);

    return (
        <div className="apis-list-wrapper">
            {loading ? <Loader /> :
                <React.Fragment>
                    <div className="apis-list-titles">
                        <div>API</div>
                        <div>{title}</div>
                    </div>
                    {
                        !!error ? null : data.map(item => (
                            <div className="apis-list-item-wrapper" key={item[apiIdKey]}>
                                <div className="api-name" onClick={() => history.push(getLink(item))}>{item.apiHostName}</div>
                                <DataDisplay {...item} />
                            </div>
                        ))
                    }
                </React.Fragment>
            }
        </div>
    );
}

export default ApisList;