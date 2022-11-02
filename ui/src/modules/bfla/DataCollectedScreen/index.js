import React, { useState } from 'react'
import classNames from 'classnames';
import IconWithTitle from 'components/IconWithTitle';
import ListDisplay from 'components/ListDisplay';
import { ICON_NAMES } from 'components/Icon';
import ModalConfirmation from 'components/ModalConfirmation';
import MessageImageDisplay from 'layout/Inventory/InventoryDetails/MessageImageDisplay';
import emptySelectImage from 'utils/images/select.svg';
import AudienceView from '../AudienceView';
import StartDetectionResumeLearningScreen from '../StartDetectionResumeLearningScreen';
import ListItemDisplayBFLA from '../ListItemDisplayBFLA';
import ToggleButtonBFLA from '../ToggleButtonBFLA';
import utils from '../utils'

import './data-collected-screen.scss'
import BreadcrumbsDisplay from 'components/BreadcrumbSelectPanes/BreadcrumbsDisplay';

export default function DataCollectedScreen({
    data,
    handleReset,
    handleMarkAsLegitimate,
    handleMarkAsIlegitimate,
    handleStartDetection,
    handleStartLearning,
}) {
    const { operations } = data;
    const collectedData = utils.formatDataForDisplay(operations);

    const [violationsOnlyState, setViolationsOnlyState] = useState(false);

    const [isViewModal, setIsViewModal] = useState(false);
    const [selectedLevelIndex, setSelectedLevelIndex] = useState(0);
    const [selectedData, setSelectedData] = useState({});

    const selectedLevelData = selectedData[selectedLevelIndex];
    let wrapperLevelData = selectedData[selectedLevelIndex === 0 ? 0 : selectedLevelIndex - 1];

    if (selectedLevelIndex === 2) {
        const matchingTagIndex = collectedData?.tags.findIndex((tagData) => tagData.name === selectedData[0].name)
        const matchingPathIndex = collectedData?.tags[matchingTagIndex]?.paths.findIndex((pathData) => pathData.path === selectedData[1].path && pathData.method === selectedData[1].method)

        wrapperLevelData = collectedData?.tags[matchingTagIndex]?.paths[matchingPathIndex]
    }

    const displayData = [
        {
            getTitle: () => "Tags",
            getSelectItems: () => collectedData && utils.getDataSelectElements(collectedData, 'tags'),
            itemDisplay: ({ name, authorized }) => <ListItemDisplayBFLA name={name} isLegitimate={authorized} />,
            checkAdvanceLevel: () => true
        },
        {
            getTitle: ({ name }) => name,
            getSelectItems: data => data && utils.getDataSelectElements(data, 'paths'),
            itemDisplay: ({ path, authorized, method }) => <ListItemDisplayBFLA method={method} name={path} isLegitimate={authorized} />,
            checkAdvanceLevel: () => true
        },
        {
            getTitle: ({ path, method }) => `${method} ${path}`,
            getSelectItems: data => data && utils.getDataSelectElements(data, 'audience'),
            itemDisplay: ({ k8s_object: { name, namespace }, authorized }) => <ListItemDisplayBFLA name={name} namespace={namespace} isLegitimate={authorized} />,
            checkAdvanceLevel: () => false
        },
    ]
    const { getTitle, checkAdvanceLevel, getSelectItems, itemDisplay: ItemDisplay } = displayData[selectedLevelIndex];

    const advanceLevel = checkAdvanceLevel(wrapperLevelData);

    return (
        <div className={classNames("spec-select-panes-wrapper-bfla")}>
            <div className="select-pane-bfla">
                {selectedLevelIndex > 0 && (
                    <BreadcrumbsDisplay
                        mainTitle="Tags"
                        selectedData={selectedData}
                        displayData={displayData}
                        selectedLevelIndex={selectedLevelIndex}
                        setSelectedLevelIndex={(selectedLevel) => {
                            setSelectedLevelIndex(selectedLevel)
                        }}
                        setSelectedData={(levelData) => {
                            setSelectedData(levelData)
                        }
                        }
                    />
                )
                }
                <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                }}>
                    <div className="select-pane-title">{getTitle(wrapperLevelData)}</div>
                    {
                        selectedLevelIndex === 0 ?
                            <div style={{
                                paddingRight: '20px'
                            }}>
                                <IconWithTitle
                                    reverseOrder
                                    onClick={() => setIsViewModal(true)}
                                    title="Reset Model"
                                    name={ICON_NAMES.ERROR}
                                />
                            </div> :
                            <div style={{
                                padding: '20px'
                            }}>
                                {
                                    getSelectItems(wrapperLevelData)
                                        .filter((item) => !item.authorized)?.length > 0 &&
                                    <ToggleButtonBFLA
                                        title="Violations only"
                                        icons={false}
                                        checked={violationsOnlyState}
                                        onChange={setViolationsOnlyState}
                                        withBoldTitle={true}
                                        small={true}
                                    />
                                }
                            </div>
                    }
                </div>
                <div className="select-items-wrapper">
                    {
                        collectedData && <ListDisplay
                            items={violationsOnlyState ? getSelectItems(wrapperLevelData).filter((tag) => tag.authorized === false) : getSelectItems(wrapperLevelData)}
                            itemDisplay={ItemDisplay}
                            selectedId={!!selectedLevelData ? selectedLevelData.id : null}
                            selectUpdatesArrow
                            onSelect={selected => {
                                setSelectedData({ ...selectedData, [selectedLevelIndex]: selected });
                                if (advanceLevel) {
                                    setSelectedLevelIndex(selectedLevelIndex + 1);
                                }
                            }}
                        />
                    }
                </div>
            </div>
            <div className="display-pane-bfla">
                {
                    selectedLevelIndex === 0 &&
                    !selectedLevelData &&
                    <StartDetectionResumeLearningScreen
                        handleStartDetection={handleStartDetection}
                        handleStartLearning={handleStartLearning}
                    />
                }
                {
                    selectedLevelIndex > 0 &&
                    !selectedLevelData &&
                    <MessageImageDisplay
                        image={emptySelectImage}
                        message={(
                            <div style={{
                                textAlign: 'center'
                            }}>
                                <div>Select an element to see details.</div>
                            </div>
                        )}
                    />
                }
                {
                    !advanceLevel &&
                    !!selectedLevelData &&
                    <AudienceView
                        selectedData={selectedData}
                        handleMarkAs={selectedData[selectedLevelIndex].authorized ? handleMarkAsIlegitimate : handleMarkAsLegitimate}
                        {...selectedData[selectedLevelIndex]}
                    />
                }
                {
                    isViewModal && <ModalConfirmation
                        title="Reset Model"
                        message={`This action will result in the deletion of all the existing data.`}
                        confirmTitle="RESET MODEL"
                        onCancle={() => setIsViewModal(false)}
                        onConfirm={handleReset}
                    />
                }
            </div>
        </div >
    )
}
