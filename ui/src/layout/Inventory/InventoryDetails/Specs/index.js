import React, { useEffect, useState, useCallback } from 'react';
import { useHistory, useRouteMatch, useLocation } from 'react-router-dom';
import classnames from 'classnames';
import { isEmpty, isNull } from 'lodash';
import { useNotificationDispatch, showNotification } from 'context/NotificationProvider';
import { useFetch, FETCH_METHODS, usePrevious } from 'hooks';
import ListDisplay from 'components/ListDisplay';
import Button from 'components/Button';
import Arrow, { ARROW_NAMES } from 'components/Arrow';
import MethodTag from 'components/MethodTag';
import Loader from 'components/Loader';
import Modal from 'components/Modal';
import BoldText from 'components/BoldText';
import UploadSpec from './UploadSpec';
import MethodHitCount from './MethodHitCount';
import { SPEC_TYPES } from './utils';

import emptySelectImage from 'utils/images/select.svg';

import './specs.scss';

const NotSelected = ({ title }) => (
    <div className="not-selected-wrapper">
        <div className="not-selected-title">{title}</div>
        <img src={emptySelectImage} alt="no tag selected" />
    </div>
);

const BackHeader = ({ title, onBack }) => (
    <div className="selected-back-header">
        <Arrow name={ARROW_NAMES.LEFT} onClick={onBack} />
        <div>{title}</div>
    </div>
);

const MethodTitle = ({ method, path }) => (
    <div className="method-item-title"><MethodTag method={method} /><span>{path}</span></div>
);

const SelectedMethodDisplay = ({ method, path, pathId, specType, inventoryName, onBack }) => (
    <div className="selected-method-wrapper">
        <BackHeader title={<MethodTitle method={method} path={path} />} onBack={onBack} />
        <MethodHitCount method={method} pathId={pathId} spec={inventoryName} specType={specType} />
    </div>
)

const SelectedTagDisplay = ({ onBack, data, inventoryName, specType }) => {
    const { name, methodAndPathList } = data;

    const [selectedMethodData, setSelectedMethodData] = useState(null);

    if (!!selectedMethodData) {
        return (
            <SelectedMethodDisplay {...selectedMethodData} specType={specType} inventoryName={inventoryName} onBack={() => setSelectedMethodData(null)} />
        )
    }

    return (
        <div className="tag-selected-wrapper">
            <BackHeader title={name} onBack={onBack} />
            <div className="tag-selected-methods-list">
                <div className="methods-list-title">Methods list</div>
                <ListDisplay
                    items={methodAndPathList.map(item => ({ ...item, id: `${item.method}${item.path}` }))}
                    itemDisplay={({ method, path }) => <MethodTitle method={method} path={path} />}
                    selectedId={!!selectedMethodData ? selectedMethodData.id : null}
                    onSelect={selected => setSelectedMethodData(selected)}
                />
            </div>
        </div>
    );
}

const SpecDisplay = ({ tags, notSelectedTitle, inventoryName, specType }) => {
    const [selectedTagData, setSelectedTagData] = useState(null);

    const tagItems = tags.map(tag => ({ id: tag.name, ...tag }));

    return (
        <div className="spec-display-wrapper">
            <div className="select-pane">
                <ListDisplay
                    items={tagItems}
                    itemDisplay={({ name }) => <span>{name}</span>}
                    selectedId={!!selectedTagData ? selectedTagData.id : null}
                    onSelect={selected => setSelectedTagData(selected)}
                />
            </div>
            <div className="display-pane">
                {isEmpty(selectedTagData) ? <NotSelected title={notSelectedTitle} /> :
                    <SelectedTagDisplay data={selectedTagData} onBack={() => setSelectedTagData(null)} inventoryName={inventoryName} specType={specType} />}
            </div>
        </div>
    )
}

const ViewInSwaggerLink = ({ inventoryId, specType }) => (
    <a href={`${window.location.origin}/swagger?apiId=${inventoryId}&specType=${specType}`} target="_blank" rel="noopener noreferrer">
        see on Swagger
    </a>
);

const ProvidedSpecDisplay = ({ specData, inventoryId, inventoryName, refreshData, specType, onReset }) => {
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
            notSelectedTitle={<span>Select a tag to see details, <ViewInSwaggerLink inventoryId={inventoryId} specType="provided" />,<br /><Button tertiary onClick={() => setShowUploadSpec(true)}>replace</Button> or <Button tertiary onClick={onReset}>reset spec</Button></span>}
            specType={specType}
        />
    )
}

const ReconstructedSpecDisplay = ({ specData, inventoryId, inventoryName, specType, onReset }) => {
    const history = useHistory();
    const { url } = useRouteMatch();

    if (!specData) {
        return (
            <NotSelected
                title={<span><Button tertiary onClick={() => history.push({ pathname: "/reviewer", query: { inventoryId, inventoryName, backUrl: url } })}>Review</Button> reconstructed spec</span>}
            />
        )
    }

    return (
        <SpecDisplay
            inventoryName={inventoryName}
            tags={specData.tags || []}
            notSelectedTitle={<span>Select a tag to see details,<br /><ViewInSwaggerLink inventoryId={inventoryId} specType="reconstructed" /> or <Button tertiary onClick={onReset}>reset</Button></span>}
            specType={specType}
        />
    )
}

export const SPEC_TAB_ITEMS = {
    PROVIDED: {
        value: SPEC_TYPES.PROVIDED,
        label: "Provided",
        dataKey: "providedSpec",
        component: ProvidedSpecDisplay,
        resetUrlSuffix: "providedSpec",
        resetConfirmationText: "Resetting the provided spec will result in loss of the uploaded spec."
    },
    RECONSTRUCTED: {
        value: SPEC_TYPES.RECONSTRUCTED,
        label: "Reconstructed",
        dataKey: "reconstructedSpec",
        component: ReconstructedSpecDisplay,
        resetUrlSuffix: "reconstructedSpec",
        resetConfirmationText: "Resetting the reconstructed spec will result in loss of spec and the information that was used to reconstruct it. If reset, to reconstruct again generate the relevant API traffic and review."
    }
}

const InnerTabs = ({ selected, items, onSelect }) => (
    <div className="spec-inner-tabs-wrapper">
        {
            items.map(({ value, label }) => (
                <div key={value} className={classnames("inner-tab-item", { selected: selected === value })} onClick={() => onSelect(value)}>{label}</div>
            ))
        }
    </div>
);

const Specs = ({ inventoryId, inventoryName }) => {
    const { query } = useLocation();
    const { inititalSelectedTab = SPEC_TAB_ITEMS.PROVIDED.value } = query || {};

    const [selectedTab, setSelectedTab] = useState(inititalSelectedTab);
    const { component: TabContentComponent, dataKey: specDataKey, value: type } = SPEC_TAB_ITEMS[selectedTab];

    const specUrl = `apiInventory/${inventoryId}/specs`;

    const [{ loading, data, error }, fetchSpecsData] = useFetch(specUrl);

    const [resetSpecType, setResetSpecType] = useState(null);
    const closeResetConfimrationodal = () => setResetSpecType(null);
    const { resetUrlSuffix, resetConfirmationText, label: resetTitle } = SPEC_TAB_ITEMS[resetSpecType] || {};

    const notificationDispatch = useNotificationDispatch();
    const showResetNotification = useCallback(() => showNotification(notificationDispatch, {
        message: <span>The <BoldText>{`${resetTitle.toLowerCase()} spec`}</BoldText> was <BoldText>reset</BoldText>.</span>
    }), [resetTitle, notificationDispatch]);


    const [{ loading: resetting, error: resetError }, resetSpecData] = useFetch(specUrl, { loadOnMount: false });
    const prevResetting = usePrevious(resetting);
    const doSpecReset = () => resetSpecData({
        formatUrl: url => `${url}/${resetUrlSuffix}`,
        method: FETCH_METHODS.DELETE
    })

    useEffect(() => {
        if (prevResetting && !resetting && !resetError) {
            showResetNotification();
            closeResetConfimrationodal();
            fetchSpecsData();
        }
    }, [prevResetting, resetting, resetError, fetchSpecsData, showResetNotification]);

    if (!!error) {
        return null;
    }

    return (
        <React.Fragment>
            <div className="inventory-details-spec-wrapper">
                {(loading || resetting) ? <Loader /> :
                    <React.Fragment>
                        <InnerTabs selected={selectedTab} items={Object.values(SPEC_TAB_ITEMS)} onSelect={selected => setSelectedTab(selected)} />
                        <TabContentComponent
                            specData={data[specDataKey]}
                            inventoryId={inventoryId}
                            inventoryName={inventoryName}
                            refreshData={fetchSpecsData}
                            specType={type}
                            onReset={() => setResetSpecType(selectedTab)}
                        />
                    </React.Fragment>
                }
            </div>
            {!isNull(resetSpecType) &&
                <Modal
                    title={`Reset ${resetTitle.toLowerCase()} spec`}
                    onClose={closeResetConfimrationodal}
                    className="spec-reset-confirmation-modal"
                    height={230}
                    onDone={() => {
                        doSpecReset();
                    }}
                    doneTitle="Reset"
                >
                    <div>{resetConfirmationText}</div>
                    <br />
                    <div>Are you sure you want to reset?</div>
                </Modal>

            }
        </React.Fragment>
    )
}

export default Specs;