import json
import yaml
import time
import base64
import requests
import logging
import threading
import os
import copy

CONFIGS = None

class Common:
    def __init__(self, TruncatedBody, body, headers, version):
        self.version = version
        self.headers = headers
        self.body = body
        self.TruncatedBody = TruncatedBody

class Request:
    def __init__(self, common, host, method, path):
        self.method = method
        self.path = path
        self.host = host
        self.common = common

class Response:
    def __init__(self, common, statusCode):
        self.statusCode = statusCode
        self.common = common

class Telemetry:
    def __init__(self, destinationAddress, destinationNamespace, request, requestID, response, scheme, sourceAddress):
        self.requestID = requestID
        self.scheme = scheme
        self.destinationAddress = destinationAddress
        self.destinationNamespace = destinationNamespace
        self.sourceAddress = sourceAddress
        self.request = request
        self.response = response

def getCommon (headers, body):
    headersList = []
    for header in headers:
        headerValue = base64.b64decode(header).decode("UTF-8")
        headerValuePair = headerValue.split(":")
        if (len(headerValuePair)>1):
            header = { "key":headerValuePair[0], "value":headerValuePair[1]};
            headersList.append(header)

    payload = base64.b64decode(body).decode("UTF-8")
    truncatedBody = False
    if (len(payload)>1000*1000):
        truncatedBody = True
        payload = ""
    version = "0.0.1"
    return Common (truncatedBody, body, headersList, version).__dict__

def getRequest (common, host, method, path):
    return Request (common, host, method, path).__dict__

def getResponse (common, statusCode):
    return Response (common, statusCode).__dict__
    
def getTelemetry (destinationAddress, request, requestID, response, scheme, sourceAddress):
    return Telemetry (destinationAddress, "", request, requestID, response, scheme, sourceAddress).__dict__

def telemetryRequestProcessing (messageJSON, token, hostsLocation, allowedhostsLocation, apiClarityURL):
    destination = messageJSON["destination"]
    requestID = messageJSON["requestID"]
    source = messageJSON["source"]
    scheme = messageJSON["scheme"]

    requestHost = messageJSON["requestHost"]
    requestMethod = messageJSON["requestMethod"]
    requestPath = messageJSON["requestPath"]
    responseStatus = messageJSON["responseStatus"]

    requestHeaders = messageJSON["requestheaders"]
    requestPayload = messageJSON["requestpayload"]
    responseHeaders = messageJSON["responseheaders"]
    responsePayload = messageJSON["responsepayload"]

    updateDiscoveredHost (destination, hostsLocation)
    if (isAllowed (destination, allowedhostsLocation)):
        logging.info("sending traces : " + destination)
        requestCommon = getCommon (requestHeaders,requestPayload)
        request = getRequest (requestCommon, requestHost, requestMethod, requestPath)

        responseCommon = getCommon (responseHeaders,responsePayload)
        response = getResponse (responseCommon, responseStatus)

        telemetry = getTelemetry (destination, request, requestID, response, scheme, source)
        telemetryJSON = json.dumps(telemetry)
        headers = getHeaders (token)
        response = requests.post(apiClarityURL, data=telemetryJSON, headers=headers)

        logging.info("response : " + response)

def isAllowed (destination, allowedhostsLocation):
    allowed = False
    with open(allowedhostsLocation, 'r') as hostFile:
        for line in hostFile:
            if (line.strip()==destination.strip()):
                allowed = True
            if (line.strip()=="*"):
                allowed = True
    return allowed
               
def updateDiscoveredHost (destination, hostsLocation):
    isNewHost = True
    with open(hostsLocation, 'r') as hostFile:
        for line in hostFile:
            if (line.strip()==destination.strip()):
                isNewHost = False
    if (isNewHost):
        with open(hostsLocation, 'a') as hostFile:
            hostFile.write(destination+"\n")

def processAPIClarity (message, token, hostsLocation, allowedhostsLocation, apiClarityURL):
    messageJSON = json.loads (message)
    telemetryRequestProcessing (messageJSON, token, hostsLocation, allowedhostsLocation, apiClarityURL)

def updateSubmittedRecord (record, recordLocation):
    with open(recordLocation, 'a') as recordFile:
        recordFile.write(record+"\n")

def isNotSubmittedRecord (record, recordLocation):
    with open(recordLocation, 'r') as recordFile:
       for line in recordFile:
            if (line.strip()==record.strip()):
                return False
    return True

def getHeaders (token):
    headers = {
        "Content-Type" : "application/json",
        "Accept" : "application/json",
        "X-Trace-Source-Token" : token.encode("UTF-8")
    }
    return headers

def updateAllowedList (urlHostToTrace, token, allowedhostsLocation):
    headers = getHeaders (token)
    response = requests.get(urlHostToTrace, headers=headers, verify=False)
    response.raise_for_status ()
    jsonResponse = response.json()
    print (jsonResponse["hosts"])
    if (len(jsonResponse["hosts"])>0):
        with open(allowedhostsLocation, 'w') as allowedHostsFile:
            for host in jsonResponse["hosts"]:
                allowedHostsFile.write(host+"\n")

def sendDiscoveredHost (urlHostList, token, hostsLocation):
    logging.info("sending sendDiscoveredHost")
    hostArray = []
    with open(hostsLocation, 'r') as hostFile:
        for line in hostFile:
            hostArray.append(line.strip())
    payload = {
        "hosts" : json.dumps(hostArray)
    }
    headers = getHeaders (token)
    response = requests.post(urlHostList, data=payload, headers=headers, verify=False)
    logging.info(response)

def apiclarity_scheduler_task():
    logging.info("Started APIClarityScheduler ...")


    apiClaritySamplerManagerURL = CONFIGS["apiclarity-url"] + "/api/hostsToTrace"
    apiClarityUpdatedHostListURL = CONFIGS["apiclarity-url"] + "/api/control/newDiscoveredAPIs"
    token = CONFIGS['apiclarity-token']
    hostsLocation = CONFIGS['hosts-path']
    allowedhostsLocation = CONFIGS['allowed-hosts-path']
    refreshInterval = CONFIGS['allowed-hosts-path']

    while True:
        try:
            updateAllowedList(apiClaritySamplerManagerURL, token, allowedhostsLocation)
            sendDiscoveredHost(apiClarityUpdatedHostListURL, token, hostsLocation)
            time.sleep(int(refreshInterval))
        except Exception as e:
            logging.error(e)

def apiclarity_processor_task():
    logging.info("Started APIClarityProcessor ...")

    apiClarityURL = CONFIGS['apiclarity-url'] + "/api/telemetry"
    token = CONFIGS['apiclarity-token']
    recordLocation = CONFIGS['record-path']
    hostsLocation = CONFIGS['hosts-path']
    allowedhostsLocation = CONFIGS['allowed-hosts-path']

    #Create files if missing
    create_file(recordLocation)
    create_file(hostsLocation)
    create_file(allowedhostsLocation)

    certfile = open(CONFIGS['apiclarity-cert-path'])

    # Config Map
    syslogFile = open('/var/log/messages', 'r')
    while syslogFile:
        try:
            line = syslogFile.readline()
            if "APICLARITY" in line:
                headerData = line.split("@")
                if (len(headerData)>2):
                    if (isNotSubmittedRecord(headerData[1], recordLocation)):
                        logging.info("processing messages ... ")
                        processAPIClarity (headerData[2], token, hostsLocation, allowedhostsLocation, apiClarityURL)
                        updateSubmittedRecord(headerData[1], recordLocation)
                        time.sleep(1)
        except Exception as e:
               logging.error(e)
    syslogFile.close()

def create_file(file):
    try:
        folder = os.path.dirname(file)
        os.makedirs(folder, exist_ok=True)
        open(file, "a").close()
    except Exception as e:
        raise Exception(f"Failed creating file {file}: {e}")


def merge_config(config:dict, overwrite:dict):
    """
    overwrite values if present values overwrite config values
    :param dict1:
    :param dict2:
    :return:
    """
    for k in overwrite:
        if isinstance(overwrite[k], dict):
            if k in config:
                merge_config(config[k], overwrite[k])
                continue

        config[k] = overwrite[k]


def config_obfuscate(config):
    c = copy.deepcopy(config)
    c['apiclarity-token'] = '**************'
    return c


def get_configs():
    # Load defaults
    with open('./config.yaml', 'r') as f:
        configs = yaml.safe_load(f) or {}

    config_filename = os.environ.get('CONFIG_PATH', None)
    if config_filename:
        try:
           with open(config_filename, 'r') as f:
                custom_configs = yaml.safe_load(f) or {}
                merge_config(configs, custom_configs)
        except Exception as e:
            logging.error(f"Unable to open config file {config_filename}: {e}")


    logging.basicConfig(level=logging.DEBUG if configs['debug'] else logging.INFO,
                            format='%(asctime)s %(levelname)s %(message)s')

    logging.debug("DEBUG IS ON")
    logging.info("\n##################################### CONFIG #####################################\n"
                 + yaml.dump(config_obfuscate(configs)) +
                 "##################################################################################")

    fatal = False
    if not configs.get('apiclarity-url'):
        logging.error("Missing config `apiclarity-url`")
        fatal = True
    if not configs.get('apiclarity-token'):
        logging.error("Missing config `apiclarity-token`")
        fatal = True

    if fatal:
        raise Exception("Invalid configuration")

    return configs

def main():
    global CONFIGS
    CONFIGS = get_configs()

    schedulerThread = threading.Thread(target=apiclarity_scheduler_task)
    schedulerThread.start()

    processorThread = threading.Thread(target=apiclarity_processor_task)
    processorThread.start()

    schedulerThread.join()
    processorThread.join()


if __name__ == "__main__":
    main()
