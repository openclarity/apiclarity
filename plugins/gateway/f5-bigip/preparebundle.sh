
#!/bin/bash

BUNDLE_FOLDER="F5BigIPBundle"

if [ -d "${BUNDLE_FOLDER}" ]; then
    echo "${BUNDLE_FOLDER} exists. Remove or delete it and then run the script again"
    exit 1
fi

if [ -z "${APICLARITY_URL}" ]; then
    echo "ERROR: APICLARITY_URL env variable not defined"
    exit 1
fi

if [ -z "${APICLARITY_TOKEN}" ]; then
    echo "ERROR: APICLARITY_TOKEN env variable not defined"
    exit 1
fi

if [ -z "${APICLARITY_CERT_PATH}" ]; then
    echo "ERROR: APICLARITY_CERT_PATH env variable not defined"
    exit 1
fi

if [ ! -f "${APICLARITY_CERT_PATH}" ]; then
    echo "ERROR: cert ${APICLARITY_CERT_PATH} cannot be found"
    exit 1
fi


cp -r F5BigIPBundle.template ${BUNDLE_FOLDER}
cp ${APICLARITY_CERT_PATH} ${BUNDLE_FOLDER}/ApiClarityAgent/deploy/apiclarity.crt
cat ${BUNDLE_FOLDER}/APIClarityAgent/deploy/config.yaml.template | sed "s|{{APIClarityURL}}|${APICLARITY_URL}|g"  | sed "s|{{APIClarityToken}}|${APICLARITY_TOKEN}|g" > ${BUNDLE_FOLDER}/APIClarityAgent/deploy/config.yaml




