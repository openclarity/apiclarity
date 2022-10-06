import React, { useState } from 'react'

import Icon, { ICON_NAMES } from 'components/Icon';
import Button from 'components/Button';
import ToggleButton from 'components/ToggleButton';
import BreadCrumbDisplayInventory from '../BreadCrumbDisplayInventory';
import ClientView from 'modules/bfla/ClientView';
import ListDisplayItem from './ListDisplayItem';

import './list-display-select.scss'

export default function ListDisplaySelect({ data }) {
    const [selectedIndex, setSelectedIndex] = useState(0);
    const [indexSelectionArr, setIndexSelectionArr] = useState([]);
    const [currentDisplayData, setCurrentDisplayData] = useState(data.tags);
    const [violationsOnlyState, setViolationsOnlyState] = useState(false);
    const [breadCrumbTitleArray, setBreadCrumbTitleArray] = useState([]);

    const handleOnIndexSelect = (indexSelected) => {
        const copyOfBreadCrumbTitleArray = [...breadCrumbTitleArray]
        copyOfBreadCrumbTitleArray.splice(indexSelected)
        setBreadCrumbTitleArray(copyOfBreadCrumbTitleArray)
        setSelectedIndex(indexSelected)
    }

    return (
        <div className='list-display-select-wrapper'>
            <div className='list-display-select-left-pane'>
                <BreadCrumbDisplayInventory
                    selectedData={breadCrumbTitleArray}
                    onIndexSelect={(indexSelected) => handleOnIndexSelect(indexSelected)}
                />
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <div className='list-display-select-title'>Title</div>
                    <div style={{
                        marginRight: '20px'
                    }}>
                        <ToggleButton
                            title={<div >Violations only</div>}
                            icons={false}
                            checked={violationsOnlyState}
                            onChange={setViolationsOnlyState}
                            withBoldTitle={true}
                            small={true}
                            disabled={false}
                        />
                    </div>
                </div>
                <Button
                    tertiary
                    icon={ICON_NAMES.ADD}
                >
                    Add new client
                </Button>

                {
                    currentDisplayData.map(({ name, isLegitimate }, index) => (
                        <ListDisplayItem
                            key={`${name}${index}`}
                            title={name}
                            icon={
                                <Icon
                                    className={isLegitimate ? "legitimate" : "illegitimate"}
                                    name={isLegitimate ? ICON_NAMES.SHIELD_CHECK : ICON_NAMES.SHIELD_CROSS}
                                />
                            }
                            onSelect={(index) => {
                                if (selectedIndex === 0) {
                                    setBreadCrumbTitleArray((value) => [...value, 'Tags'])
                                    setCurrentDisplayData(data.tags[index])
                                } else if (selectedIndex === 1) {
                                    setBreadCrumbTitleArray((value) => [...value, 'Paths'])
                                    setCurrentDisplayData(data.tags[selectedIndex])
                                } else if (selectedIndex === 2) {
                                    setBreadCrumbTitleArray((value) => [...value, 'Paths'])
                                    setCurrentDisplayData(data.tags[selectedIndex])
                                }
                            }}
                        />
                    )
                    )
                }
            </div>
            <div className='list-display-select-right-pane'>
                <ClientView
                    isEditMode={false}
                    selectedData={{
                        name: "client view",
                        isLegitimate: false,
                        principles: [
                            // {
                            //     principleType: "Basic Auth",
                            //     name: "Hoe Doe",
                            //     ip: '10.22.33.44'
                            // }
                        ],
                        lastObserved: new Date(),
                        lastStatusCode: "200",
                        namespace: "prod",
                    }}
                    handleEditLegitimacyStatus={() => console.log("EDIT LEGITIMACY")}
                    handleMarkAs={() => console.log("MARK AS")}
                />
            </div>
        </div>
    )
}
