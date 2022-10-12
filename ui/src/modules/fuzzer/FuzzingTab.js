import React, { useState, useEffect } from 'react';
import { useFetch, usePrevious } from 'hooks';
import { isEmpty } from 'lodash';
import Loader from 'components/Loader';
import FormModal from 'components/FormModal';
import TestSelectPanes from './TestSelectPanes';
import NewTestForm from './NewTestForm';
import EmptyFuzzingDisplay from './EmptyFuzzingDisplay';
import { convertTestResponse } from './utils';

const TabFuzzingTest = (props) => {
    const {id, outerHistory, hasProvidedSpec, hasReconstructedSpec} = props;
    const [{loading, data, error}, fetchTests] = useFetch(`modules/fuzzer/tests/${id}`);
    const prevLoading = usePrevious(loading);

    const [showStartTestForm, setShowStartTestForm] = useState(false);
    const openStartTestForm = () => setShowStartTestForm(true);
    const closeStartTestForm = () => setShowStartTestForm(false);

    const [displayDataConfig, setDisplayDataConfig] = useState({displayData: [], initialLoadDone: false});
    const {displayData, initialLoadDone} = displayDataConfig;

    useEffect(async () => {
        if (prevLoading && !loading && !error) {
            setDisplayDataConfig({displayData: await convertTestResponse(id, data)  , initialLoadDone: true});
        }
    }, [prevLoading, loading, error, data]);

    if (!initialLoadDone) {
        return <Loader />;
    }

    if (error) {
		return null;
    }

    const isFuzzable = hasProvidedSpec || hasReconstructedSpec;

    return (
        <React.Fragment>
            {isEmpty(displayData) ?
                <EmptyFuzzingDisplay onStart={openStartTestForm} isFuzzable={isFuzzable} outerHistory={outerHistory} /> :
                <TestSelectPanes catalogId={id} testElements={displayData} onNewTestClick={openStartTestForm} onScanComplete={fetchTests} isFuzzable={isFuzzable} />
            }
            {showStartTestForm &&
                <FormModal
                    onClose={closeStartTestForm}
                    formComponent={NewTestForm}
                    formProps={{
                        catalogId: id,
                        onFormSubmitSuccess: () => {
                            closeStartTestForm();
                            fetchTests();
                        }
                    }}
                />
            }
        </React.Fragment>

    )
}

export default TabFuzzingTest;
