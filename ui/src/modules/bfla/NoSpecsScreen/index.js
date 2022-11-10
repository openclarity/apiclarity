import React from 'react'
import { useHistory } from 'react-router-dom';
import MessageImageDisplay from 'layout/Inventory/InventoryDetails/MessageImageDisplay'
import MessageTextButton from 'components/MessageTextButton'
import emptySelectImage from 'utils/images/empty_select_image.svg'

export default function NoSpecScreen({ id }) {
    const history = useHistory();

    return (
        <MessageImageDisplay
            message={
                <div style={{ textAlign: 'center' }}>
                    <MessageTextButton
                        onClick={() => history.push(`/inventory/INTERNAL/${id}`)}
                    >
                        Upload a spec
                    </MessageTextButton> or <MessageTextButton
                        onClick={() => history.push(`/inventory/INTERNAL/${id}`)}
                    >
                        reconstruct
                    </MessageTextButton> one in order to
                    <br />
                    enable BFLA detection for this API
                </div>}
            image={emptySelectImage}
        />
    )
}
