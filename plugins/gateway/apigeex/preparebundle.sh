
#!/bin/bash

BUNDLE_FILE="ApiClarityFlowBundle.zip"
BUNDLE_FOLDER="sharedflowbundle"

if [ -f "${BUNDLE_FILE}" ]; then
    echo "${BUNDLE_FILE} exists. Remove or delete it and then run the script again"
    exit 1
fi

if [ -d "${BUNDLE_FOLDER}" ]; then
    echo "${BUNDLE_FOLDER} exists. Remove or delete it and then run the script again"
    exit 1
fi

if [ -z "${APICLARITY_URL}" ]; then
    echo "ERROR: APICLARITY_URL env variable not defined"
    exit 1
fi

if [ -z "${APICLARITY_APIGEEX_TOKEN}" ]; then
    echo "ERROR: APICLARITY_URL env variable not defined"
    exit 1
fi


cp -r sharedflowbundle.template ${BUNDLE_FOLDER}
cat sharedflowbundle.template/policies/JavaScript-APIClarity-Tracing.xml | sed "s|{{APIClarityURL}}|${APICLARITY_URL}|g"  | sed "s|{{APIClarityToken}}|${APICLARITY_APIGEEX_TOKEN}|g" > ${BUNDLE_FOLDER}/policies/JavaScript-APIClarity-Tracing.xml
zip -r ${BUNDLE_FILE} sharedflowbundle

rm -rf ${BUNDLE_FOLDER}

