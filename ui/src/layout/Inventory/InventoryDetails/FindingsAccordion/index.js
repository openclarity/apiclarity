import React from 'react';
import { isNull, isEmpty, get } from 'lodash';
import { API_RISK_ITEMS } from 'utils/systemConsts';
import COLORS from 'utils/scss_variables.module.scss';
import Icon, { ICON_NAMES } from 'components/Icon';
import Text, { TEXT_TYPES } from 'components/Text';
import Accordion from 'components/Accordion';
import ColorByRisk from 'components/ColorByRisk';
import TextWithLinks from 'components/TextWithLinks';
import JsonDisplayBox from 'components/JsonDisplayBox';
import AdditionalInfoContent from 'layout/Inventory/InventoryDetails/AdditionalInfoContent';
// import DownloadJsonButton from '../DownloadJsonButton';
import DownloadJsonButton from 'components/DownloadJsonButton';

import './findings-accordion.scss';

const getFormattedSpecSubPath = (specPath, removePrefix) => specPath.replace(removePrefix, "/").replaceAll("['", "/").replaceAll("']", "").replace("//", "/");

const FindingDetails = ({ withOccurrencesCount, source, description, mitigation, additional_info, additionalInfo, data, endpoints }) => {
    const displayArr = withOccurrencesCount ? [
        { label: 'Source', value: source },
        { label: 'Description', value: <TextWithLinks text={description} /> },
        { label: 'Mitigation', value: mitigation },
    ] : [
        { label: 'Description', value: <TextWithLinks text={description} /> },
        { label: 'Additional Info', value: <AdditionalInfoContent additionalInfo={{ entries: additionalInfo }} /> }
    ];

    const checkDisplayOccurences = (additionlInfos) => {
        if (isEmpty(additionalInfo)) {
            const emptyEntriesArr = additionlInfos.filter((additionlInfo) => isEmpty(additionlInfo.entries))

            return emptyEntriesArr.length !== additionlInfos.length
        }

        return false;
    }

    if (withOccurrencesCount && checkDisplayOccurences(additional_info)) {
        displayArr.push(
            {
                label: 'Occurrences', value: (
                    additional_info.map((additionalInfoItem, index) =>
                        additionalInfoItem && !isEmpty(additionalInfoItem.entries) && (
                            <Accordion key={index} customTitle={
                                <div className='cutom-accordion-title'>
                                    {`#${index + 1} Additional Info`}
                                </div>
                            }>
                                <AdditionalInfoContent additionalInfo={additionalInfoItem} endpoints={endpoints} />
                            </Accordion>
                        )
                    )
                )
            }
        )
    }

    return (
        <div className="finding-details-wrapper">
            {withOccurrencesCount &&
                <div className="findings-actions-wrapper">
                    <div style={{ marginRight: '1rem' }}>
                        <DownloadJsonButton title="Download finding's JSON" fileName="findings-data" data={data} />
                    </div>
                </div>
            }
            {displayArr.map(({ label, value }) => (
                <div
                    key={label}
                    style={{
                        display: 'grid',
                        gridTemplateColumns: '1fr 3fr',
                        marginBlock: '1rem',
                    }}
                >
                    <div style={{
                        fontWeight: 'bold',
                        color: COLORS['color-grey-dark']
                    }}>
                        {label}
                    </div>
                    <div>
                        {value}
                    </div>
                </div>
            ))}
        </div>
    )
};

const SpacPathFindingDetails = ({ description, mitigation, specPath, specPathPrefix, specJson }) => {
    const DataDisplay = ({ title, children }) => (
        <div className="spec-finding-content-item">
            <div className="data-item-title"><Text type={TEXT_TYPES.TABLE_HEADER}>{title}</Text></div>
            <div className="data-item-content"><Text type={TEXT_TYPES.TABLE_BODY}>{children}</Text></div>
        </div>
    );

    const PathsSubPath = () => (
        <DataDisplay title="Sub path">
            {getFormattedSpecSubPath(specPath, specPathPrefix)}
        </DataDisplay>
    );

    const COMPONENTS_PATH = "['components']['schemas']";
    const COMPONENTS_PATH_PREFIX = `$${COMPONENTS_PATH}`;

    const ComponentsSubPath = () => {
        const componentName = specPath.replace(COMPONENTS_PATH_PREFIX, "").replace("['", "").split("'")[0];
        const componentPath = `${COMPONENTS_PATH}['${componentName}']`;
        const componentPathPrefix = `$${componentPath}`;

        return (
            <DataDisplay title="Affected component">
                <div style={{ marginBottom: "5px" }}>{`Subpath: ${getFormattedSpecSubPath(specPath, componentPathPrefix)}`}</div>
                <JsonDisplayBox json={{ [componentName]: get(specJson, componentPath) }} />
            </DataDisplay>
        )
    }

    return (
        <div className="spec-path-finding-details-wrapper">
            <DataDisplay title="Description"><TextWithLinks text={description} /></DataDisplay>
            <DataDisplay title="Mitigation">{mitigation}</DataDisplay>
            {!!specPath && (specPath.startsWith(COMPONENTS_PATH_PREFIX) ? <ComponentsSubPath /> : <PathsSubPath />)}
        </div>
    );
}

const FindingsAccordion = ({ findingsDetails, withSpecPath = false, elementsKey = "findings", specPathPrefix, specJson, endpoints, withOccurrencesCount = true }) => {
    const { critical, high, medium, low, unclassified } = findingsDetails || {};

    const riskLevels = [
        { level: API_RISK_ITEMS.CRITICAL.value, label: "Critical risk findings", content: critical },
        { level: API_RISK_ITEMS.HIGH.value, label: "High risk findings", content: high },
        { level: API_RISK_ITEMS.MEDIUM.value, label: "Medium risk findings", content: medium },
        { level: API_RISK_ITEMS.LOW.value, label: "Low risk findings", content: low },
        { level: API_RISK_ITEMS.NO_RISK.value, label: "No known risk findings", content: unclassified },
    ]

    const nonEmptyRiskFindings = riskLevels.filter(item => !isNull(item.content) && item.content?.count !== 0);

    const FindingsComponent = withSpecPath ? SpacPathFindingDetails : FindingDetails;

    return (
        <div className="findings-details-accordion">
            {!isEmpty(nonEmptyRiskFindings) &&
                nonEmptyRiskFindings.map((item, index) => {
                    const { level, label, content = { [elementsKey]: [], count: 0 } } = item;

                    const title = (
                        <React.Fragment>
                            <ColorByRisk risk={level} isText={false}>
                                <Icon name={ICON_NAMES.BEETLE_ROUND} />
                            </ColorByRisk>
                            <span>{content.count} {label}</span>
                        </React.Fragment>
                    )

                    return (
                        <Accordion key={index} title={title} className="risk-level-accordion" isEmpty={content.count === 0}>
                            {
                                content[elementsKey].map((finding, findingIndex) => (
                                    <Accordion key={findingIndex} customTitle={
                                        <div style={{
                                            display: 'flex',
                                        }}>
                                            <div style={{
                                                display: 'table',
                                            }}>
                                                <p style={{
                                                    display: 'table-cell',
                                                    verticalAlign: 'middle'
                                                }}>
                                                    {finding.name}
                                                </p>
                                            </div>
                                            {
                                                withOccurrencesCount &&
                                                <div style={{
                                                    color: COLORS['color-main'],
                                                    backgroundColor: COLORS['color-blue'],
                                                    textAlign: 'center',
                                                    marginLeft: '1rem',
                                                    paddingInline: '0.8rem',
                                                    paddingBlock: '0.5rem'
                                                }}>{finding.occurrences && finding.occurrences > 1 ? (`${finding.occurrences} Occurrences`) : `${finding.occurrences || 1} Occurrence`} </div>
                                            }
                                        </div >
                                    } className="risk-finding-accordion" >
                                        <FindingsComponent withOccurrencesCount={withOccurrencesCount} {...finding} specPathPrefix={specPathPrefix} specJson={specJson} endpoints={endpoints} />
                                    </Accordion >
                                ))
                            }
                        </Accordion >
                    )
                })
            }
        </div >
    )
}

export default FindingsAccordion;
