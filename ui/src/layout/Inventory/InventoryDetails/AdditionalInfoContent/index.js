import React from 'react'
import JsonDisplayBox from 'components/JsonDisplayBox'

import './additional-info-content.scss'

export default function AdditionalInfoContent({ additionalInfo, endpoints }) {
    const { affected_endpoints: affectedEndpoints, entries } = additionalInfo

    return (
        <div className='additiona-info-container'>
            {affectedEndpoints && affectedEndpoints.length > 0 && (
                <div className='affected-enpoints-container'>
                    <div className='affected-enpoints-items-container'>
                        <span className='affected-enpoints-item' >
                            Affected endpoint{affectedEndpoints.length > 1 ? 's' : ''}:&nbsp;
                        </span>
                        {affectedEndpoints?.map((affectedEndpoint, index) => {
                            const endpointObject = endpoints?.find(({ endpoint }) => {
                                return endpoint.identifier === affectedEndpoint
                            })
                            return (
                                <span key={index} className='affected-enpoints-item' >
                                    {endpointObject?.endpoint?.host}{endpointObject?.endpoint?.port ? `:${endpointObject?.endpoint?.port}` : ''}
                                    {index < affectedEndpoints.length - 1 ? ',' : ''}
                                    &nbsp;
                                </span>
                            )
                        }
                        )}
                    </div>
                </div>
            )}

            {entries && Object.keys(entries).map((entryKey) => {
                let value = '';

                try {
                    value = JSON.parse(entries[entryKey])

                    if (value === undefined || typeof value !== 'object') {
                        value = entries[entryKey]
                    }

                } catch (err) {
                    value = entries[entryKey]
                } finally {
                    return (
                        <div key={entryKey} style={{ marginBlock: '1rem' }}>
                            <div style={{
                                display: 'flex'
                            }}>
                                <div className='additional-info-key-style'>
                                    title:
                                </div>
                                <div>
                                    {entryKey}
                                </div>
                            </div>
                            <div style={{
                                display: 'flex'
                            }}>
                                <div className='additional-info-key-style'>
                                    value:
                                </div>
                                <div>
                                    {
                                        typeof value === 'string' ?
                                            value
                                            :
                                            <JsonDisplayBox style={{ padding: 0 }} json={value} />
                                    }
                                </div>
                            </div>
                        </div>
                    )
                }
            })}
        </div>
    )
}
