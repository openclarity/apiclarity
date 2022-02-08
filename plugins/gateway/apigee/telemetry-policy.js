var maxBodySize = 1000 * 1000

// REQUEST
//------>Header<------//
var requestHeaderList = [];

var requestHeaders = context.getVariable("request.headers.names"),

requestHeaders = requestHeaders + '';

requestHeaders = requestHeaders.slice(1, -1).split(', ');

requestHeaders.forEach(function(k){
  var v = context.getVariable("request.header." + k );
  var header = { key : k, value: v};
  requestHeaderList.push(header);

});
//------>Body<------//
var truncatedReqBody = false

var reqContent = context.getVariable('request.content')

if(reqContent !== null){

    if (reqContent.length > maxBodySize){
		truncatedBodyReq = true
		reqContent = ""
    }
}


var requestCommon = {
                        truncatedBody : truncatedReqBody,
                        body: crypto.base64(crypto.asBytes(reqContent)),
                        headers: requestHeaderList,
                        version: "0.0.1"
                    };

var req = {
            common : requestCommon,
            host: context.getVariable("request.header.host"),
            method: context.getVariable('request.verb'),
            path: context.getVariable("request.uri")
          };


// RESPONSE
//------>Header<------//
var responseHeaderList = [];

var responseHeaders = context.getVariable("response.headers.names"),

responseHeaders = responseHeaders + '';

responseHeaders = responseHeaders.slice(1, -1).split(', ');

responseHeaders.forEach(function(k){
  var v = context.getVariable("response.header." + k );
  var header = { key : k, value: v};
  responseHeaderList.push(header);
});

//------>Body<------//
var truncatedResBody = false

var resContent = context.getVariable('response.content')

if(resContent !== null){

    if (resContent.length > maxBodySize){
		truncatedResBody = true
		resContent = ""
    }
}

var responseCommon = {
                        truncatedBody : truncatedResBody,
                        body: crypto.base64(crypto.asBytes(resContent)),
                        headers: responseHeaderList,
                        version: "0.0.1"
                    };

var res = {
            common : responseCommon,
            statusCode: context.getVariable('response.status.code').toString()
          };


var telemetry = {
                    destinationAddress: ":"+context.getVariable("target.port"),
                    destinationNamespace: "",
                    request: req,
                    requestID: context.getVariable("request.header.x-request-id"),
                    response: res,
                    scheme: context.getVariable('client.scheme'),
                    sourceAddress: context.getVariable("request.header.x-forwarded-for")+":"
                }

print("Telemetry:", JSON.stringify(telemetry));

var url = "http://<API_CLARITY_SERVICE>:9000/api/telemetry";

var payload = JSON.stringify(telemetry);

var headers = {
                 'Content-Type' : 'application/json',
                 'Accept' : 'application/json'
              };

var req = new Request(url, 'POST', headers, payload);


httpClient.send(req, onComplete);

function onComplete(response, error) {
    if (response) {
      print(context.getVariable('response.status'))
    }
    else {
      context.setVariable('apiclarity.error', 'oops: ' + error);
    }
  }