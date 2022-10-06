import React from 'react'
import Button from 'components/Button'
import MessageImageDisplay from 'layout/Inventory/InventoryDetails/MessageImageDisplay'
import inProgressImage from 'images/in_progress.svg'

export default function DataCollectionInProgress({ id, state }) {



    return (
        <>
            <MessageImageDisplay
                message={"BFLA model learning in progress..."}
                image={inProgressImage}
            />
            <div style={{
                display: 'flex',
                justifyContent: 'center',
                paddingBottom: '30px'
            }}>
                <Button tertiary>
                    Stop BFLA model learning
                </Button>
            </div>
        </>
    )
}
