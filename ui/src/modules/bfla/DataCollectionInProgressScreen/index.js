import React from 'react'
import Button from 'components/Button'
import MessageImageDisplay from 'layout/Inventory/InventoryDetails/MessageImageDisplay'
import inProgressImage from 'utils/images/in_progress.svg'

export default function DataCollectionInProgress({ isLearning, handleStop }) {

    return (
        <>
            <MessageImageDisplay
                message={`BFLA model ${isLearning ? 'learning' : 'detecting'} in progress...`}
                image={inProgressImage}
            />
            <div style={{
                display: 'flex',
                justifyContent: 'center',
                paddingBottom: '30px'
            }}>
                <Button tertiary onClick={handleStop}>
                    Stop BFLA model {isLearning ? 'learning' : 'detecting'}
                </Button>
            </div>
        </>
    )
}
