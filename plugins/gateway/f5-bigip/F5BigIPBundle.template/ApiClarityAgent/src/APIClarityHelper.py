import base64


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


def getCommon(headers, body):
    headersList = []
    for header in headers:
        headerValue = base64.b64decode(header).decode("UTF-8")
        headerValuePair = headerValue.split(":")
        if (len(headerValuePair) > 1):
            header = {"key": headerValuePair[0], "value": headerValuePair[1]};
            headersList.append(header)

    payload = base64.b64decode(body).decode("UTF-8")
    truncatedBody = False
    if (len(payload) > 1000 * 1000):
        truncatedBody = True
        payload = ""
    version = "0.0.1"
    return Common(truncatedBody, body, headersList, version).__dict__


def getRequest(common, host, method, path):
    return Request(common, host, method, path).__dict__


def getResponse(common, statusCode):
    return Response(common, statusCode).__dict__


def prepare_telemetry(trace):
    destination = trace["destination"]
    requestID = trace["requestID"]
    source = trace["source"]
    scheme = trace["scheme"]

    requestHost = trace["requestHost"]
    requestMethod = trace["requestMethod"]
    requestPath = trace["requestPath"]
    responseStatus = trace["responseStatus"]

    requestHeaders = trace["requestheaders"]
    requestPayload = trace["requestpayload"]
    responseHeaders = trace["responseheaders"]
    responsePayload = trace["responsepayload"]

    requestCommon = getCommon(requestHeaders, requestPayload)
    request = getRequest(requestCommon, requestHost, requestMethod, requestPath)
    responseCommon = getCommon(responseHeaders, responsePayload)
    response = getResponse(responseCommon, responseStatus)

    return Telemetry(destination, "", request, requestID, response, scheme, source).__dict__
