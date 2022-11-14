import React, { useState } from 'react';
import classnames from 'classnames';
import ListDisplay from 'components/ListDisplay';
import MessageImageDisplay from 'layout/Inventory/InventoryDetails/MessageImageDisplay';
import BreadcrumbsDisplay from './BreadcrumbsDisplay';

// import emptySelectImage from 'layout/Apis/ApiInventory/images/empty_selection.svg';
import emptySelectImage from 'utils/images/select.svg';

import './breadcrumbs-select-panes.scss';

export const SelectItemNotification = ({title}) => (
    <MessageImageDisplay
        image={emptySelectImage}
        message={title}
    />
)

const BreadcrumbSelectPanes = ({mainBreadcrumbsTitle, displayData, className}) => {
    const [selectedLevelIndex, setSelectedLevelIndex] = useState(0);
    const [selectedData, setSelectedData] = useState({}); //{0: selectedLevelData, 1: undefined, 2: ...}

    const {getTitle, checkAdvanceLevel, getSelectItems, itemDisplay: ItemDisplay, selectContentDisplay: SelectContentDisplay,
        customHeaderComponent: CustomHeaderComponent, emptySelectDisplay: EmptySelectDisplay} = displayData[selectedLevelIndex];

    const selectedLevelData = selectedData[selectedLevelIndex];
    const wrapperLevelData = selectedData[selectedLevelIndex === 0 ? 0 : selectedLevelIndex - 1];

    const advanceLevel = checkAdvanceLevel(wrapperLevelData);

    return (
        <div className={classnames("breadcrumbs-select-panes-wrapper", className)}>
            <div className="select-pane">
                <BreadcrumbsDisplay
                    selectedData={selectedData}
                    displayData={displayData}
                    selectedLevelIndex={selectedLevelIndex}
                    setSelectedLevelIndex={setSelectedLevelIndex}
                    setSelectedData={setSelectedData}
                    mainTitle={mainBreadcrumbsTitle}
                />
                <div className="select-pane-title">{getTitle(wrapperLevelData)}</div>
                {!!CustomHeaderComponent &&
                     <div className="select-pane-custom-title"><CustomHeaderComponent /></div>
                }
                <div className="select-items-wrapper">
                    <ListDisplay
                        items={getSelectItems(wrapperLevelData)}
                        itemDisplay={ItemDisplay}
                        selectedId={!!selectedLevelData ? selectedLevelData.id : null}
                        onSelect={selected => {
                            setSelectedData({ ...selectedData, [selectedLevelIndex]: selected });

                            if (advanceLevel) {
                                setSelectedLevelIndex(selectedLevelIndex + 1);
                            }
                        }}
                        selectUpdatesArrow
                    />
                </div>
            </div>
            <div className="display-pane">
                {!advanceLevel && !!selectedLevelData && <SelectContentDisplay {...selectedLevelData} />}
                {!selectedLevelData && !!EmptySelectDisplay && <EmptySelectDisplay {...wrapperLevelData} />}
            </div>
        </div>
    )
}

export default BreadcrumbSelectPanes;
