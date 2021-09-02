import React, { useState } from 'react';
import { useHistory, useRouteMatch, useLocation } from 'react-router-dom';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import { useFetch } from 'hooks';
import ListDisplay from 'components/ListDisplay';
import Button from 'components/Button';
import Arrow, { ARROW_NAMES } from 'components/Arrow';
import Tag from 'components/Tag';
import Loader from 'components/Loader';
import UploadSpec from './UploadSpec';
import MethodHitCount from './MethodHitCount';
import emptySelectImage from 'utils/images/select.svg';

import './specs.scss';

const NotSelected = ({title}) => (
    <div className="not-selected-wrapper">
        <div className="not-selected-title">{title}</div>
        <img src={emptySelectImage} alt="no tag selected" />
    </div>
);

const BackHeader = ({title, onBack}) => (
    <div className="selected-back-header">
        <Arrow name={ARROW_NAMES.LEFT} onClick={onBack} />
        <div>{title}</div>
    </div>
);

const MethodTitle = ({method, path}) => (
    <div className="method-item-title"><Tag>{method}</Tag><span>{path}</span></div>
);

const SelectedMethodDisplay = ({method, path, inventoryName, onBack}) => (
    <div className="selected-method-wrapper">
        <BackHeader title={<MethodTitle method={method} path={path} />} onBack={onBack} />
        <MethodHitCount method={method} path={path} spec={inventoryName} />
    </div>
)

const SelectedTagDisplay = ({onBack, data, inventoryName}) => {
    const {name, methodAndPathList} = data;

    const [selectedMethodData, setSelectedMethodData] = useState(null);

    if (!!selectedMethodData) {
        return (
            <SelectedMethodDisplay {...selectedMethodData} inventoryName={inventoryName} onBack={() => setSelectedMethodData(null)} />
        )
    }

    return (
        <div className="tag-selected-wrapper">
            <BackHeader title={name} onBack={onBack} />
            <div className="tag-selected-methods-list">
                <div className="methods-list-title">Methods list</div>
                <ListDisplay
                    items={methodAndPathList.map(item => ({...item, id: `${item.method}${item.path}`}))}
                    itemDisplay={({method, path}) => <MethodTitle method={method} path={path} />}
                    selectedId={!!selectedMethodData ? selectedMethodData.id : null}
                    onSelect={selected => setSelectedMethodData(selected)}
                />
            </div>
        </div>
    );
}

const SpecDisplay = ({tags, notSelectedTitle, inventoryName}) => {
    const [selectedTagData, setSelectedTagData] = useState(null);

    const tagItems = tags.map(tag => ({id: tag.name, ...tag}));

    return (
        <div className="spec-display-wrapper">
            <div className="select-pane">
                <ListDisplay
                    items={tagItems}
                    itemDisplay={({name}) => <span>{name}</span>}
                    selectedId={!!selectedTagData ? selectedTagData.id : null}
                    onSelect={selected => setSelectedTagData(selected)}
                />
            </div>
            <div className="display-pane">
                {isEmpty(selectedTagData) ? <NotSelected title={notSelectedTitle} /> :
                    <SelectedTagDisplay data={selectedTagData} onBack={() => setSelectedTagData(null)} inventoryName={inventoryName} />}
            </div>
        </div>
    )
}

const ViewInSwaggerLink = ({inventoryId, specType}) => (
    <a href={`${window.location.origin}/swagger?apiId=${inventoryId}&specType=${specType}`} target="_blank" rel="noopener noreferrer">
        view in swagger
    </a>
);

const ProvidedSpecDisplay = ({specData, inventoryId, inventoryName, refreshData}) => {
    const [showUploadSpec, setShowUploadSpec] = useState(!specData);

    if (showUploadSpec) {
        return (
            <UploadSpec title="Upload spec" onUpdate={refreshData} inventoryId={inventoryId} />
        )
    }

    return (
        <SpecDisplay
            inventoryName={inventoryName}
            tags={specData.tags || []}
            notSelectedTitle={<span>Select a tag to see details, <ViewInSwaggerLink inventoryId={inventoryId} specType="provided" /> or <Button secondary onClick={() => setShowUploadSpec(true)}>replace spec</Button></span>}
        />
    )
}

const ReconstructedSpecDisplay = ({specData, inventoryId, inventoryName}) => {
    const history = useHistory();
    const {url} = useRouteMatch();

    if (!specData) {
        return (
            <NotSelected
                title={<span><Button secondary onClick={() => history.push({pathname: "/reviewer", query: {inventoryId, inventoryName, backUrl: url}})}>Review</Button> reconstructed spec</span>}
            />
        )
    }

    return (
        <SpecDisplay
            inventoryName={inventoryName}
            tags={specData.tags || []}
            notSelectedTitle={<span>Select a tag to see details or <ViewInSwaggerLink inventoryId={inventoryId} specType="reconstructed" /></span>}
        />
    )
}

export const SPEC_TAB_ITEMS = {
    PROVIDED: {value: "PROVIDED", label: "Provided", dataKey: "providedSpec", component: ProvidedSpecDisplay},
    RECONSTRUCTED: {value: "RECONSTRUCTED", label: "Reconstructed", dataKey: "reconstructedSpec", component: ReconstructedSpecDisplay}
}

const InnerTabs = ({selected, items, onSelect}) => (
    <div className="spec-inner-tabs-wrapper">
        {
            items.map(({value, label}) => (
                <div key={value} className={classnames("inner-tab-item", {selected: selected === value})} onClick={() => onSelect(value)}>{label}</div>
            ))
        }
    </div>
);

const Specs = ({inventoryId, inventoryName}) => {
    const {query} = useLocation();
    const {inititalSelectedTab=SPEC_TAB_ITEMS.PROVIDED.value} = query || {};
    
    const [selectedTab, setSelectedTab] = useState(inititalSelectedTab);
    const {component: TabContentComponent, dataKey: specDataKey} = SPEC_TAB_ITEMS[selectedTab];

    const [{loading, data, error}, fetchSpecsData] = useFetch(`apiInventory/${inventoryId}/specs`);

    if (!!error) {
        return null;
    }

    return (
        <div className="inventory-details-spec-wrapper">
            {loading ? <Loader /> : 
                <React.Fragment>
                    <InnerTabs selected={selectedTab} items={Object.values(SPEC_TAB_ITEMS)} onSelect={selected => setSelectedTab(selected)} />
                    <TabContentComponent specData={data[specDataKey]} inventoryId={inventoryId} inventoryName={inventoryName} refreshData={fetchSpecsData} />
                </React.Fragment>
            }
        </div>
    )
}

export default Specs;