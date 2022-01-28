// REQUEST
//------>Header<------//
var requestHeaderList = [];

var requestHeaders = context.getVariable("request.headers.names"),

requestResults = {};
    
requestHeaders = requestHeaders + '';

requestHeaders = requestHeaders.slice(1, -1).split(', ');

requestHeaders.forEach(function(k){
  var v = context.getVariable("request.header." + k );
  requestResults[k] = v;
  var header = { key : k, value: v};
  requestHeaderList.push(header);
  
});
//------>Body<------//
var reqContent = context.getVariable('request.content')

if(reqContent !== null){
    reqContent = crypto.base64(crypto.asBytes(reqContent))
}


var requestCommon = { 
                        truncatedBody : true, 
                        body: reqContent, 
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

responseResults = {};

responseHeaders = responseHeaders + '';

responseHeaders = responseHeaders.slice(1, -1).split(', ');

responseHeaders.forEach(function(k){
  var v = context.getVariable("response.header." + k );
  responseResults[k] = v;
  var header = { key : k, value: v};
  responseHeaderList.push(header);
});

//------>Body<------//
var resContent = context.getVariable('response.content')

if(resContent !== null){
    resContent = crypto.base64(crypto.asBytes(resContent))
}

var responseCommon = { 
                        truncatedBody : true, 
                        body: resContent, 
                        headers: responseHeaderList, 
                        version: "0.0.1"
                    };
                    
var res = { 
            common : responseCommon,
            statusCode: context.getVariable('response.status.code').toString()
          };


var telemetry = {
                    destinationAddress: ":80",
                    destinationNamespace: "",
                    request: req,
                    requestID: context.getVariable("request.header.x-request-id"),
                    response: res,
                    scheme: context.getVariable('client.scheme'),
                    sourceAddress: ":80"
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



