import Button from 'components/Button'
import React from 'react'

import './data-collection-bfla.scss'

const DATA_COLLECTION_SCREEN_TITLE = 'BFLA model learning'
const DATA_COLLECTION_SCREEN_SUBTITLE = 'Sending and analysing traces uses cluster resources (network traffic, CPU consumtion, storage consumptions)'

export default function DataCollectionScreen({ handleStartModelLearning }) {
    return (
        <div className='data-collection-container-bfla' >
            <div className='data-collection-title-bfla' >
                {DATA_COLLECTION_SCREEN_TITLE}
            </div>
            <div className='data-collection-subtitle-bfla' >
                {DATA_COLLECTION_SCREEN_SUBTITLE}
            </div>
            <Button
                onClick={handleStartModelLearning}
            >
                START
            </Button>
        </div>
    )
}
