import React from 'react'
import Arrow, { ARROW_NAMES } from 'components/Arrow'

import './breadcrumb-display-inventory.scss'

export default function BreadCrumbDisplayInventory({ selectedData, onIndexSelect }) {

    const handleBreadCrumbClick = (selectedDataIndex) => selectedDataIndex !== selectedData.length - 1 ?
        onIndexSelect(selectedDataIndex) :
        undefined

    return selectedData && (
        <div className='breadcrumb-display-inventory-wrapper'>
            {selectedData.map((selectedDataTitle, selectedDataIndex) => (
                <div
                    key={`${selectedDataTitle}${selectedDataIndex}`}
                    style={{ display: 'flex' }}
                >
                    <div
                        className='breadcrumb-display-inventory-item'
                        onClick={() => handleBreadCrumbClick(selectedDataIndex)}
                    >
                        {selectedDataTitle}
                    </div>
                    {
                        selectedDataIndex !== selectedData.length - 1 && <div style={{
                            display: 'flex',
                            justifyContent: 'center',
                            alignItems: 'center'
                        }}>
                            <Arrow
                                name={ARROW_NAMES.RIGHT}
                                small
                            />
                        </div>
                    }
                </div>
            ))}
        </div>
    )
}
