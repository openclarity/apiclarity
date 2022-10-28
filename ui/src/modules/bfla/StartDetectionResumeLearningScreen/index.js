import React from 'react'
import MessageImageDisplay from 'layout/Inventory/InventoryDetails/MessageImageDisplay'
import MessageTextButton from 'components/MessageTextButton'
import emptySelectImage from 'utils/images/empty_select_image.svg'

export default function StartDetectionResumeLearningScreen({ handleStartDetection, handleStartLearning }) {
    return (
        <MessageImageDisplay
            message={
                <div style={{ textAlign: 'center' }}>
                    Select a tag to see methods
                    <br />
                    <MessageTextButton
                        onClick={handleStartDetection}
                    >
                        Start BFLA detection
                    </MessageTextButton> or <MessageTextButton
                        onClick={handleStartLearning}
                    >
                        resume BFLA model learning
                    </MessageTextButton>
                </div>
            }
            image={emptySelectImage}
        />
    )
}
