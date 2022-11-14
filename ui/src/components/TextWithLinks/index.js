import React from 'react';

//Transforms text with format: "___LINK___Displayed Text___HREF___url___LINK___" into a tag links

const LINK_MARK = "___LINK___";
const HREF_MARK = "___HREF___";
const FORMATTING_PLACEHOLDER = "___LINKPLACEHOLDER___";

const TextWithLinks = ({text}) => {
    let placeholderText = text;

    const replaceTextArrays = [...text.matchAll(LINK_MARK + "(.*?)" + LINK_MARK)];

    const linksInText = replaceTextArrays.map(([fullLinkText, innerLinkText]) => {
        placeholderText = placeholderText.replace(fullLinkText, FORMATTING_PLACEHOLDER);

        return innerLinkText;
    });

    return (
        placeholderText.split(FORMATTING_PLACEHOLDER).map((textItem, index) => {
            const linksInTextItem = linksInText[index];

            if (!linksInTextItem) {
                return textItem;
            }

            const [displayText, url] = linksInTextItem.split(HREF_MARK);

            return (
                <span key={index}>{textItem}<a href={url} target="_blank" rel="noopener noreferrer">{displayText}</a></span>
            );
        })
    )
}

export default TextWithLinks;
