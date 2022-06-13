import React from 'react';
import { useHistory } from 'react-router-dom';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import StatusIndicator from 'components/StatusIndicator';
import MethodTag from 'components/MethodTag';
import Button from 'components/Button';

const Details = ({data}) => {
    const {method, statusCode, path, query, sourceIP, destinationIP, destinationPort, hostSpecName, apiInfoId, apiType} = data;

    const history = useHistory();

    return (
        <div>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Method"><MethodTag method={method} /></TitleValueDisplay>
                <TitleValueDisplay title="Status code"><StatusIndicator title={statusCode} isError={statusCode >= 400} /></TitleValueDisplay>
            </TitleValueDisplayRow>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Path" className="path-display">{path}</TitleValueDisplay>
            </TitleValueDisplayRow>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Query" className="query-display">{query}</TitleValueDisplay>
            </TitleValueDisplayRow>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Source">{sourceIP}</TitleValueDisplay>
                <TitleValueDisplay title="Destination">{destinationIP}</TitleValueDisplay>
            </TitleValueDisplayRow>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Destination port">{destinationPort}</TitleValueDisplay>
            </TitleValueDisplayRow>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="spec" className="spec-display">
                    {!!apiInfoId ? <Button secondary onClick={() => history.push(`/inventory/${apiType}/${apiInfoId}`)}>{hostSpecName}</Button> : hostSpecName}
                </TitleValueDisplay>
            </TitleValueDisplayRow>
        </div>
    )
}

export default Details;
