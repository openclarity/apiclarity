var maxBodySize = String (properties.maxMessageSize);
var refreshInterval = String (properties.refreshInterval);
var apiClarityUrl = String (properties.apiClarityUrl);
var urlTrace = apiClarityUrl + "/api/telemetry";
var urlHostToTrace = apiClarityUrl + "/api/hostsToTrace";
var urlToUpdateHostList = apiClarityUrl + "/api/control/newDiscoveredAPIs";

var token = String (properties.apiclarityToken)
var cacheUpdatedTime = context.getVariable("cacheUpdatedTime");
var hostList = context.getVariable("hostList");
var allowedHostList = context.getVariable("allowedHostList");
var destinationAddr = context.getVariable("target.host") + ":" + context.getVariable("target.port");

var currentTime = new Date().getTime();
var refreshed = isRefresh (refreshInterval, cacheUpdatedTime, currentTime);
context.setVariable('APIClarity.refreshed', refreshed);

if (refreshed === true) {
    context.setVariable('cacheUpdatedTimeValue', currentTime);
    updateHosts (destinationAddr, token, urlToUpdateHostList, hostList);
}
var allowed = isAllowed (refreshed, allowedHostList, token, urlHostToTrace, destinationAddr);
context.setVariable('APIClarity.allowed', allowed);

if (allowed === true){
    var payload = getTelemetryPayload (maxBodySize);
    var headers = getHTTPHeader (token);
    // Next line is for debugging purposes. 
    // context.setVariable('APIClarity.Trace.payload', payload);
    
    var req = new Request(urlTrace, 'POST', headers, payload);
    httpClient.send(req, onCompleteTrace);
}

function onCompleteTrace(response, error){
    if (response) {
        context.setVariable('APIClarity.Trace.success', response.status);
    } else {
        context.setVariable('APIClarity.Trace.error', error);
    }
}


function getHTTPHeader (token){
    var headers = {
        'Content-Type' : 'application/json',
        'Accept' : 'application/json',
        'X-Trace-Source-Token' : token
      };
    return headers;
}

function getHeaders (contextName){
    var headerList = [];
    var headers = context.getVariable(contextName + ".headers.names");
   
    if (headers!=null){
        headers = String(headers);
        headers = headers.slice(1, -1).split(", ");
        headers.forEach(function(hName){
            if (hName!=null && (String(hName)!=""))
            var hValue = context.getVariable(contextName + ".header." + String(hName).trim());
            if (hValue!=null){
                var header = { key : String(hName).trim(), value: hValue};
                headerList.push(header);
            }
        });
    } else if (contextName=="request"){
        headerList.push( {key :"request.header.host", value: context.getVariable("request.header.host")});
        headerList.push( {key :"request.header.x-forwarded-for", value: context.getVariable("request.header.x-forwarded-for")});
        headerList.push( {key :"request.header.x-request-id", value: context.getVariable("request.header.x-request-id")});
        headerList.push( {key :"request.header.content-type", value: context.getVariable("request.header.content-type")});
    }
    return headerList;
}

function getCommon (contextName, maxBodySize){
    var truncatedReqBody = false;
    var content = context.getVariable(contextName + ".content");
    if(content !== null){
        if (content.length > maxBodySize){
		    truncatedBodyReq = true;
		    content = "";
        }
    }
    var common = {
        truncatedBody : truncatedReqBody,
        body: crypto.base64(crypto.asBytes(content)),
        headers: getHeaders(contextName),
        version: "0.0.1"
    };
    return common;
}

function getRequest (maxBodySize){
    var request = {
        common : getCommon("request", maxBodySize),
        host: context.getVariable("request.header.host"),
        method: context.getVariable('request.verb'),
        path: context.getVariable("request.uri")
      };
    return request;
}

function getResponse (maxBodySize){
    var response = {
        common : getCommon("response", maxBodySize),
        statusCode: context.getVariable('response.status.code').toString()
      };
    return response;
}

function getTelemetryPayload (maxBodySize){
    var telemetry = {
        destinationAddress:  destinationAddr,
        destinationNamespace: "",
        request: getRequest(maxBodySize),
        requestID: context.getVariable("request.header.x-request-id"),
        response: getResponse(maxBodySize),
        scheme: context.getVariable('client.scheme'),
        sourceAddress: context.getVariable("request.header.x-forwarded-for")+":"
    };
    var payload = JSON.stringify(telemetry);
    return payload;
}

function isRefresh (refreshInterval, cacheUpdatedTime, currentTime){
    if (cacheUpdatedTime===null || cacheUpdatedTime==="" || cacheUpdatedTime==="null"){
        return true;
    }
    try{
        var lastUpdatedTime = Number (cacheUpdatedTime);
        if ( (currentTime-lastUpdatedTime)<refreshInterval){
            return false;
        }
    }
    catch (err){
    }
    return true;
}

function updateHosts (destinationAddr, token, urlToUpdateHostList, hostList){
    var hostListArr = populateHostList (destinationAddr, hostList);
    sendHostList(token, urlToUpdateHostList, hostListArr);
}

function populateHostList (destinationAddr, hostList){
    var hostListArr = [];
    var hostStr = destinationAddr;
    if (hostList!=null && hostList!="null"){
        var found = false;
        hostAddressses=hostList.split(",");
        if ( hostAddressses.length > 0 ){
                hostAddressses.forEach(function(item){
                    hostListArr.push(String(item));
                if (String(destinationAddr)===String(item)){
                    found = true; 
                }
            });
        }
        if (found===false){
            hostStr = hostList + "," + destinationAddr;
        }
    }
    hostListArr.push(destinationAddr);
    context.setVariable("hostListValue", hostStr);
    return hostListArr;
}

function getHostListPayload (hosts){
    var payload = {
        hosts : hosts
    }
    return JSON.stringify(payload);
}

function sendHostList (token, urlToUpdateHostList, hostListArr){
    var headers = getHTTPHeader (token);
    var payload = getHostListPayload (hostListArr);
    context.setVariable('APIClarity.SendHosts.request', payload);
    var req = new Request (urlToUpdateHostList, 'POST', headers, payload);
    httpClient.send(req, onCompleteHost);
}

function onCompleteHost (){
    if (response) {
        context.setVariable('APIClarity.SendHosts.success', response.status);
    } else {
        context.setVariable('APIClarity.SendHosts.error', error);
    }
}

function isAllowed (refreshed, allowedHostList, token, urlHostToTrace, destinationAddr){
    var allowedList = allowedHostList;
    if (refreshed===true){
        allowedList = updateAllowedHostList (token, urlHostToTrace);
    }

    var allowed = false;
    var allowedListArr = allowedList.split(",");
    if ( allowedListArr.length > 0 ){
        allowedListArr.forEach(function(item){ 
            if ("*"===String(item)){
                allowed = true; 
            }
            else if (String(item)===String(destinationAddr)){
                allowed = true; 
            }
            else {
                hostPort = item.split(":");
                if (hostPort.length===1){
                    destinationHostPort = destinationAddr.split(":");
                    if (String(hostPort[0])===String(destinationHostPort[0])){
                        allowed = true;
                    }
                }
            }
        });
        return allowed;
    }
}

function updateAllowedHostList (token, urlHostToTrace){
    var headers = getHTTPHeader (token);
    var req = new Request(urlHostToTrace, 'GET', headers);
    var exchange = httpClient.send(req);
    exchange.waitForComplete();

    var allowedHostList = '';
    if (exchange.isSuccess()) {
        var responseObj = exchange.getResponse().content.asJSON;
        if (responseObj.hosts!==null && responseObj.hosts!=="null"){
                responseObj.hosts.forEach(function (k){
                if (allowedHostList===''){
                allowedHostList = String(k);
                } else {
                 allowedHostList = allowedHostList + "," + String(k);
                }
            });
        }
        else{
            allowedHostList = "*";
        }
        context.setVariable("allowedHostListValue", allowedHostList);

    } else if (exchange.isError()) {
        context.setVariable('APIClarity.Sampler.error', exchange.getError());
    }
    return allowedHostList;
}