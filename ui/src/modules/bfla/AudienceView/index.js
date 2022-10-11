import React from 'react';
import classnames from 'classnames'
import { isEmpty } from 'lodash'
import DisplaySection from 'components/DisplaySection';
import Icon, { ICON_NAMES } from 'components/Icon';
import Tooltip from 'components/Tooltip';
import Button from 'components/Button';
import TableSimple from 'components/TableSimple';
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

const ClientViewPrincipelsView = ({ end_users }) => {
    return (
        end_users ?
            <TableSimple
                headers={["Type", "Principal", "IP"]}
                name="principals"
                rows={end_users.map(({ source, id, ip_address }) =>
                ([
                    <div style={{ lineHeight: '60px' }}>{source}</div>,
                    id,
                    ip_address,
                ]))}
            /> :
            <></>
    )
}

const ClientViewTitle = ({ name, namespace, authorized, handleMarkAsClick }) => (
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
            onClick={handleMarkAsClick}
        >
            Mark as {authorized ? 'Illegitimate' : 'Legitimate'}
        </Button>
    </div>
)

export default function AudienceView({ handleMarkAs, selectedData, authorized, end_users, external, k8s_object, lastTime, statusCode, warningStatus }) {
    const { namespace, name, uid } = k8s_object;

    const handleMarkAsClick = () => handleMarkAs(selectedData[1].method, selectedData[1].path, uid)

    return (
        <div className='client-view-wrapper'>
            {
                name && namespace &&
                <DisplaySection title={
                    <ClientViewTitle handleMarkAsClick={handleMarkAsClick} authorized={authorized} name={name} namespace={namespace} />
                } >
                    <ClientViewBasicInfo
                        lastObserved={lastTime}
                        lastStatusCode={statusCode}
                        isLegitimate={authorized}
                    />
                </DisplaySection >
            }
            {
                end_users && !isEmpty(end_users) && <DisplaySection title={"Principals"} >
                    <ClientViewPrincipelsView end_users={end_users} />
                </DisplaySection>
            }
        </div>
    )
}
