import React, { useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import { isEmpty } from 'lodash';
import Loader from 'components/Loader';
import NoResultsDisplay from 'components/NoResultsDisplay';
import { useFetch } from 'hooks';

import './apis-list.scss';

const ApisList = ({url, getLink, apiIdKey, refreshTimestamp, columnItems}) => {
    const history = useHistory();

    const [{loading, data, error}, fetchData] = useFetch(url);

    useEffect(() => {
        fetchData();
    }, [fetchData, refreshTimestamp]);

    const cellWidth = `${100 / columnItems.length + 1}%`;
    const getCellStyle = (index, items=[]) => ({width: cellWidth, ...(items.length - 1 === index ? {textAlign: "end"} : {})});

    return (
        <div className="apis-list-wrapper">
            {loading ? <div style={{marginTop: "50px"}}><Loader /></div> :
                (isEmpty(data) ? <NoResultsDisplay title="No results found" isSmall /> :
                    <React.Fragment>
                        <div className="apis-list-titles">
                            <div style={getCellStyle()}>API</div>
                            {columnItems.map(({title}, index, items) => <div key={title} style={getCellStyle(index, items)}>{title}</div>)}
                        </div>
                        {
                            !!error ? null : data.map((item) => (
                                <div className="apis-list-item-wrapper" key={item[apiIdKey]}>
                                    <div className="api-name" style={getCellStyle()} onClick={() => history.push(getLink(item))}>
                                        {item.apiHostName}
                                    </div>
                                    {columnItems.map(({title, content: Content}, index, items) =>
                                        <div key={title} style={getCellStyle(index, items)}><Content {...item} /></div>)}
                                </div>
                            ))
                        }
                    </React.Fragment>
                )
            }
        </div>
    );
}

export default ApisList;