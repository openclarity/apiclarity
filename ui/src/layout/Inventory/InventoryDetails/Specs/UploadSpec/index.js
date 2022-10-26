import React, { useEffect, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { isNull } from 'lodash';
import classNames from 'classnames';
import { useFetch, FETCH_METHODS, usePrevious } from 'hooks';
import Icon, { ICON_NAMES } from 'components/Icon';
import Button from 'components/Button';

import './upload-spec.scss';

const NoFileSepected = ({ getRootProps, getInputProps, onBrowse }) => (
    <div {...getRootProps({ className: "dropzone-wrapper" })}>
        <div className="inner-wrapper">
            <input {...getInputProps()} />
            <Icon name={ICON_NAMES.DOWNLOAD} />
            <div>{`Drag file here or `}<Button tertiary onClick={onBrowse}>browse</Button></div>
        </div>
    </div>
);

const FileSelected = ({ fileName, onRemove, disabled }) => (
    <div className="file-selected-wrapper">
        <Icon name={ICON_NAMES.DOWNLOAD_JSON} className="file-icon" />
        <div className="file-name">{fileName}</div>
        <Icon
            name={ICON_NAMES.X_MARK}
            onClick={() => disabled ? null : onRemove()}
            className="close-icon"
            disabled={disabled}
        />
    </div>
);

const readFile = (file) => {
    return new Promise((resolve) => {
        const reader = new FileReader();
        reader.onerror = () => console.log('error reading file');
        reader.onload = () => {
            resolve(reader.result);
        };
        reader.readAsText(file);
    });
};

const UploadSpec = ({ title, inventoryId, onUpdate }) => {
    const [{ error, loading }, uploadSpec] = useFetch(`apiInventory/${inventoryId}/specs/providedSpec`, { loadOnMount: false });
    const prevLoading = usePrevious(loading);

    const [selectedFile, setSlectedFile] = useState(null);
    const noFileSelected = isNull(selectedFile);

    useEffect(() => {
        if (prevLoading && !loading && !error) {
            onUpdate();
        }
    }, [prevLoading, loading, error, onUpdate]);

    const onUploadSpec = async () => {
        const fileToUpload = await readFile(selectedFile);

        uploadSpec({ submitData: { rawSpec: fileToUpload }, method: FETCH_METHODS.PUT });
    };

    const { getRootProps, getInputProps, open: onBrowse } = useDropzone({
        onDrop: acceptedFiles => setSlectedFile(acceptedFiles[0] || null),
        multiple: false,
        noClick: true,
        noKeyboard: true
    });

    return (
        <div className="upload-spec-wrapper">
            <div className="upload-spec-title">{title}</div>
            <div className="upload-spec-content">
                <div className="upload-spec-file-container">
                    {noFileSelected ?
                        <NoFileSepected getRootProps={getRootProps} getInputProps={getInputProps} onBrowse={onBrowse} /> :
                        <FileSelected fileName={selectedFile.name} onRemove={() => setSlectedFile(null)} disabled={loading} />
                    }
                </div>
                <div className="upload-spec-file-container-footer">
                    {(loading || !!error) &&
                        <div className={classNames("upload-spec-file-status-message", { error: !loading && !!error })}>
                            {loading ? "Processing..." : "Error in file"}
                        </div>
                    }
                    <div className="upload-spec-file-supported-message">Supported formats are OAS v2.0 and OAS v3.0 (json or yaml)</div>
                </div>
            </div>
            <div className="submit-button-wrapper">
                <Button disabled={loading || noFileSelected} onClick={onUploadSpec} className="file-submit-button">Finish</Button>
            </div>
        </div>
    )
}

export default UploadSpec;