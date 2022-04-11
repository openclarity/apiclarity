import React, { useState } from 'react';
import classnames from 'classnames';
import SpecDiff from './SpecDiff';

const SPEC_TYPES = {
    PROVIDED_DIFF: "PROVIDED_DIFF",
    RECONSTRUCTED_DIFF: "RECONSTRUCTED_DIFF"
};

const SPEC_TAB_ITEMS = {
    PROVIDED_DIFF: {
        value: SPEC_TYPES.PROVIDED_DIFF,
        label: "Provided",
        dataKey: "providedSpecDiff",
        component: ({url}) => <SpecDiff url={url}/>,
        urlSuffix: "providedSpecDiff",
    },
    RECONSTRUCTED_DIFF: {
        value: SPEC_TYPES.RECONSTRUCTED_DIFF,
        label: "Reconstructed",
        dataKey: "reconstructedSpecDiff",
        component: ({url}) => <SpecDiff url={url}/>,
        urlSuffix: "reconstructedSpecDiff",
    }
};

const InnerTabs = ({selected, items, onSelect}) => (
    <div className="spec-inner-tabs-wrapper">
        {
            items.map(({value, label}) => (
                <div key={value} className={classnames("inner-tab-item", {selected: selected === value})} onClick={() => onSelect(value)}>{label}</div>
            ))
        }
    </div>
);

const Specs = ({data}) => {
    const [selectedTab, setSelectedTab] = useState(SPEC_TAB_ITEMS.PROVIDED_DIFF.value);
    const {component: TabContentComponent, urlSuffix} = SPEC_TAB_ITEMS[selectedTab];
    const {id: eventId} = data;

    return (
        <div className="events-spec-wrapper">
            <React.Fragment>
                <InnerTabs selected={selectedTab} items={Object.values(SPEC_TAB_ITEMS)} onSelect={selected => setSelectedTab(selected)} />
                <TabContentComponent url={`apiEvents/${eventId}/${urlSuffix}`} />
            </React.Fragment>
        </div>
    );
};

export default Specs;
