import React from 'react';
import classnames from 'classnames'
import { isEmpty } from 'lodash'
import DisplaySection from 'components/DisplaySection';
import Icon, { ICON_NAMES } from 'components/Icon';
import Tooltip from 'components/Tooltip';
import Button from 'components/Button';
import Table from 'components/Table';
import { formatDate } from 'utils/utils';
import COLORS from 'utils/scss_variables.module.scss'

import './client-view.scss';


const LEGITAMCY_STATUS_OBJECT = {
    LEGITIMATE: {
        label: "Legitimate",
        value: true,
    },
    ILLEGITIMATE: {
        label: "Illegitimate",
        value: false,
    }
}

const DATA_TIP_ID = '0'

const ClientViewBasicInfo = ({
    isLegitimate,
    lastObserved,
    lastStatusCode,
}) => {
    return (
        <div className='bfla-client-vire-basic-info'>
            <div>
                <div className='bfla-client-title'>
                    STATUS
                </div>
                <div
                    data-tip
                    data-for={DATA_TIP_ID}
                    style={{
                        display: 'flex',
                        alignItems: 'center'
                    }}
                >
                    <Icon
                        className={classnames('bfla-status-icon', isLegitimate ? "legitimate" : "illegitimate")}
                        name={
                            isLegitimate ?
                                ICON_NAMES.SHIELD_CHECK :
                                ICON_NAMES.SHIELD_CROSS
                        }
                    />
                    <div
                        className='bfla-client-value'
                    >
                        {isLegitimate ? LEGITAMCY_STATUS_OBJECT.LEGITIMATE.label : LEGITAMCY_STATUS_OBJECT.ILLEGITIMATE.label}
                    </div>
                </div>
                {!isLegitimate &&
                    <Tooltip
                        id={DATA_TIP_ID}
                        text={
                            <span>
                                Potential Broken Function Level Authorisation
                                <br />
                                call violating the current authorization model.
                                <br />
                                Please verify authorisation implementation in
                                <br />
                                the API server.
                            </span>
                        }
                    />
                }
            </div>
            <div>
                <div className='bfla-client-title'>
                    LAST OBSERVED
                </div>
                <div className='bfla-client-value'>
                    {lastObserved ? formatDate(lastObserved) : "-"}
                </div>
            </div>
            <div>
                <div className='bfla-client-title'>
                    LAST RESPONSE CODE
                </div>
                <div className='bfla-client-value'>
                    {lastStatusCode ? lastStatusCode : "-"}
                </div>
            </div>
        </div >
    )
}

const ClientViewPrincipelsView = ({ principals }) => {

    // TODO fix the table
    const enhancedPrincipals = principals.map((data) => ({ ID: "sadsdasd", ...data }))
    return (<></>)
}

export default function ClientView({ selectedData, isEditMode, handleEditLegitimacyStatus, handleMarkAs }) {
    const { name, isLegitimate, principles, lastObserved, lastStatusCode, namespace } = selectedData;

    const handleInputChange = (changeData) => {
        handleEditLegitimacyStatus(selectedData, changeData.target.value === "false" ? false : true)
    }

    return (
        <div className='client-view-wrapper'>
            {
                name && namespace &&
                <DisplaySection title={
                    <div style={{
                        display: 'flex'
                    }}>
                        <div style={{
                            display: 'flex',
                            width: '100%',
                            fontSize: '14px',
                            letterSpacing: "0.36px",
                        }}>
                            <div>
                                {name}
                            </div>
                            <div style={{
                                fontWeight: 'normal',
                                color: COLORS['color-grey-dark'],
                            }}>
                                &nbsp;| Namespace: {namespace}
                            </div>
                        </div>
                        <Button
                            tertiary
                            onClick={() => handleMarkAs(isLegitimate)}
                        >
                                Mark as {isLegitimate ? 'Illegitimate' : 'Legitimate'}
                        </Button>
                    </div>
                } >
                    <ClientViewBasicInfo
                        isEditMode={isEditMode}
                        lastObserved={lastObserved}
                        lastStatusCode={lastStatusCode}
                        isLegitimate={isLegitimate}
                        handleInputChange={handleInputChange}
                    />
                </DisplaySection >
            }
            {
                principles && !isEmpty(principles) && <DisplaySection title={"Principals"} >
                    <ClientViewPrincipelsView principals={principles} />
                </DisplaySection>
            }
        </div>
    )
}
