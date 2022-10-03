import React from 'react';
import MessageImageDisplay from 'layout/Inventory/InventoryDetails/MessageImageDisplay';
import MessageTextButton from 'components/MessageTextButton';
// import { DEPLOYMNET_CLUSTERS_PAGE_PATH } from 'layout/Deployments/Clusters';
import Button from 'components/Button';

import newTestImage from 'utils/images/add_new.svg';
import updateClusterImage from 'utils/images/select.svg';

import './empty-fuzzing-display.scss';

const EmptyFuzzingDisplay = ({isFuzzable, onStart, clusterName, outerHistory}) => {
    if (!isFuzzable) {
        return (
            // <MessageImageDisplay
            //     className="fuzzing-disabled-message-wrapper"
            //     message={(
            //         <div className="fuzzing-disabled-message">
            //             <div className="fuzzing-disabled-message-title">Fuzzing disabled</div>
            //             <div className="fuzzing-disabled-message-content">
            //                 Your cluster configurations don't allow fuzzing options.
            //                 <Button tertiary onClick={() => outerHistory.push({pathname: DEPLOYMNET_CLUSTERS_PAGE_PATH, query: {openFormClusterName: clusterName}})}>Edit the settings</Button> to enable testing
            //             </div>
            //         </div>
            //     )}
            //     image={updateClusterImage}
            // />
            <MessageImageDisplay
                className="fuzzing-disabled-message-wrapper"
                message={(
                    <div className="fuzzing-disabled-message">
                        <div>Upload a spec</div>
                        {/* <div className="fuzzing-disabled-message-title">Fuzzing disabled</div> */}
                        {/* <div className="fuzzing-disabled-message-content"> */}
                        {/*     Your cluster configurations don't allow fuzzing options. */}
                        {/*     <Button tertiary onClick={() => outerHistory.push({pathname: DEPLOYMNET_CLUSTERS_PAGE_PATH, query: {openFormClusterName: clusterName}})}>Edit the settings</Button> to enable testing */}
                        {/* </div> */}
                    </div>
                )}
                image={updateClusterImage}
            />
        )
    }

    return (
        <MessageImageDisplay message={<MessageTextButton onClick={onStart}>Start new fuzz test</MessageTextButton>} image={newTestImage} />
    )
}

export default EmptyFuzzingDisplay;
