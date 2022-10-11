import React from 'react'
import classNames from 'classnames';
import Arrow, { ARROW_NAMES } from 'components/Arrow';

const Breadcrumb = ({ title, onClick, hideArrow = false }) => (
    <div className="breadcrumb-wrapper">
        {!hideArrow && <Arrow name={ARROW_NAMES.RIGHT} small />}
        <div className={classNames("breadcrumb-title", { clickable: !!onClick })} onClick={onClick}>{title}</div>
    </div>
);

export default function BreadcrumbsDisplayBFLA({ selectedData, displayData, selectedLevelIndex, setSelectedLevelIndex, setSelectedData }) {
    const nonEmptySelectedDataKeys = Object.keys(selectedData).filter(level => !!selectedData[level]);

    return (
        <div className="select-breadcrumbs">
            {nonEmptySelectedDataKeys.length > 0 &&
                <Breadcrumb
                    title="Tags"
                    onClick={() => {
                        setSelectedLevelIndex(0);
                        setSelectedData({});
                    }}
                    hideArrow={true}
                />
            }
            {
                nonEmptySelectedDataKeys.map((level) => {
                    const selectedLevelData = selectedData[level];
                    const levelInt = parseInt(level);
                    const levelIntWithOffset = levelInt + 1; //level 0 is outside this mapping (offsetting count)
                    const { checkAdvanceLevel } = displayData[levelInt];

                    if (!selectedLevelData) {
                        return null;
                    }

                    const onLevelClick = () => {
                        const updatedSelectedData = Object.keys(selectedData).map(key => {
                            const data = selectedData[key];
                            const intKey = parseInt(key);

                            return (intKey <= levelInt) ? data : null;
                        })

                        setSelectedLevelIndex(levelIntWithOffset);
                        setSelectedData(updatedSelectedData);
                    }

                    return (
                        checkAdvanceLevel(selectedData[levelInt - 1]) &&
                        <Breadcrumb key={level} title={selectedLevelData.name || `${selectedLevelData.method} ${selectedLevelData.title}`} onClick={levelIntWithOffset === selectedLevelIndex ? undefined : onLevelClick} />
                    )
                })
            }
        </div>
    )
}