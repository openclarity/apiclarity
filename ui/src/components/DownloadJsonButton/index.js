import React from 'react';
import { ICON_NAMES } from 'components/Icon';
import IconWithTitle from 'components/IconWithTitle';

const DownloadJsonButton = ({title, fileName, data}) => {
    const downloadFile = () => {
        const file = new Blob([JSON.stringify(data, null, 2)], {type: "text/plain"});

        const element = document.createElement("a");
        element.href = URL.createObjectURL(file);
        element.download = `${fileName}.json`;
        element.click();
    };

    return (
        <IconWithTitle name={ICON_NAMES.DOWNLOAD_JSON} title={title} onClick={downloadFile} />
    );
};

export default DownloadJsonButton;
