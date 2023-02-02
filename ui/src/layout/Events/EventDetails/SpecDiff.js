import React from 'react';
import ReactDiffViewer from 'react-diff-viewer-continued';
import { useFetch } from 'hooks';
import Loader from 'components/Loader';

const SpecDiff = ({url}) => {
    const [{loading, data}] = useFetch(url);

    const {newSpec, oldSpec} = data || {};

    return (
        <div className="spec-diff-wrapper">
            {loading ? <Loader /> :
                <React.Fragment>
                    <div className="spec-diff-titles-wrapper">
                        {!newSpec && !oldSpec ?
                            <div>No Spec Available</div> :
                            <React.Fragment>
                                <div>Documented</div>
                                <div>Detected</div>
                            </React.Fragment>
                        }
                    </div>
                    <ReactDiffViewer oldValue={oldSpec} newValue={newSpec} splitView={true} />
                </React.Fragment>
            }
        </div>
    )
}

export default SpecDiff;
