export const SEPARATOR = "/";

export const checkIsParam = section => section.startsWith("{") && section.endsWith("}");

export const getPathWithParamInIndex = (path, index, paramName) => {
    const pathList = path.split(SEPARATOR);

    if (pathList.length <= index) {
        return path;
    }

    const updatedPathList = [
        ...pathList.slice(0, index),
        `{${paramName}}`,
        ...pathList.slice(index + 1)
    ];
    
    return updatedPathList.join(SEPARATOR);
}

export const getMethodsFromPaths = paths => (
    [...new Set(paths.reduce((acc, curr) => {
        return [...acc, ...curr.methods];
    }, []))]
)