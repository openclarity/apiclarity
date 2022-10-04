import React, { useState, useEffect } from 'react';
import moment from 'moment';
import { useFetch, usePrevious } from 'hooks';
import { isEmpty, cloneDeep } from 'lodash';
import Loader from 'components/Loader';
import FormModal from 'components/FormModal';
import TestSelectPanes from './TestSelectPanes';
import NewTestForm from './NewTestForm';
import TestDetails from './TestDetails';
import EmptyFuzzingDisplay from './EmptyFuzzingDisplay';
import { formatDateBy } from 'utils/utils';
import {asyncDataFetcher} from 'utils/apiUtils';


// current API Clarity tests response:
// {
//     "items": [
//         {
//             "errorMessage": "",
//             "progress": 100,
//             "starttime": 1662740147,
//             "vulnerabilities": {
//                 "critical": 0,
//                 "high": 4,
//                 "low": 0,
//                 "medium": 0,
//                 "total": 4
//             }
//         }
//     ],
//     "total": 1
// }

const sortTestsDescending = (tests) => {
    return tests.sort((a, b) => {
        return moment.utc(b.starttime).diff(moment.utc(a.starttime));
    });
};

const getTest = (apiId, starttime) => {
    return new Promise ((resolve) => {
        asyncDataFetcher({url: `modules/fuzzer/report/${apiId}/${starttime}/short`, successCallback: (data) => resolve(data)});
    });
};

const convertTestResponse = async (apiId, data) => {
    const testModel = {
        tags: { elements: [], severity: null },
        testDetails: {
            testId: '',
            fuzzingProgress: 0,
            fuzzingStartTime: null,
            fuzzingStatus: null,
            fuzzingStatusMessage: null,
            testConfiguration: {
                auth: null,
                depth: null
            }
        }
    };

    // Temp workaround until status is returned from backend
    const getTestStatus = (test) => {
        if (test.errorMessage) {
            return 'ERROR';
        }
        return test.progress === 100 ? 'DONE' : 'IN_PROGRESS';
    };

    let items = data.items || [];
    items = sortTestsDescending(items);
    items = await Promise.all(items.map( async (t) => {
        const report = await getTest(apiId, t.starttime);
        t.tags = report.tags || [];
        t.tags = t.tags.map((t) => {
            t.methods = t.operations.map((op) => {
                return {
                    findings: {
                        elements: op.findings
                    },
                    path: op.operation.path,
                    method: op.operation.method,
                    highestSeverity: op.highestSeverity,
                    requestCount: op.requestsCount
                };
            });

            return t;
        });
        return t;
    }));

    items = items.map((test) => {
        const model = cloneDeep(testModel);
        model.testDetails = {
            ...model.testDetails,
            ...{
                testId: '' + test.starttime,
                fuzzingStatusMessage: test.errorMessage,
                fuzzingProgress: test.progress,
                fuzzingStartTime: formatDateBy(test.starttime * 1000),

                // TODO: need status to be on the list of tests from backend.
                fuzzingStatus: getTestStatus(test),
            }};

        // model.tags.severity = 'INFO';
        model.tags.elements = test.tags;
        return model;
    });
    return items;
};

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
                            /* closeStartTestForm(); */
                            /* fetchTests(); */
                            // TODO: remove settimeout.
                            setTimeout(() => {
                                closeStartTestForm();
                                fetchTests();
                            }, 1000);
                        }
                    }}
                />
            }
        </React.Fragment>

    )
}

export default TabFuzzingTest;
