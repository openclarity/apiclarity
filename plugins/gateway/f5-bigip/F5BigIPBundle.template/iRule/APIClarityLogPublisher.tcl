when CLIENT_ACCEPTED {
    set apiclarity_hsl [HSL::open -publisher /Common/APIClarity-Log-Publisher]
    set source_address [IP::client_addr]:[TCP::client_port]
}

when HTTP_REQUEST_SEND {
    clientside {
        set destination_address "[IP::server_addr]:[TCP::server_port]"
        set request_method [HTTP::method]
        set request_path [HTTP::path]
        set request_host [HTTP::header value Host]
        if {[string first ":" $request_host] == -1} {
            set request_host "$request_host:[TCP::server_port]"
        }
        
        
        
        set contentLength [HTTP::header "Content-Length"]
        if {[HTTP::header exists "Content-Length"]}{
            HTTP::collect [HTTP::header "Content-Length"]
        } else {
            set http_request_payload_encoded ""
        }
    
        set request_index 0
        foreach aHeader [HTTP::header names] {
            if {$request_index==0} {
                set hValue [HTTP::header value $aHeader]
                set requestHeaderValue "$aHeader:$hValue"
                set requestHeaderValueEncoded [b64encode $requestHeaderValue]
                set request_headers "\"$requestHeaderValueEncoded\""
            } else {
                set hValue [HTTP::header value $aHeader]
                set hValue [HTTP::header value $aHeader]
                set requestHeaderValue "$aHeader:$hValue"
                set requestHeaderValueEncoded [b64encode $requestHeaderValue]
                set request_headers "$request_headers,\"$requestHeaderValueEncoded\""
            }
            set request_index [expr {$request_index+1}]
        }
        set request_headers "\[$request_headers\]" 
    }
}

when HTTP_REQUEST {
    set request_id [TMM::cmp_unit][clock clicks]
    set scheme "http"
}

when HTTP_REQUEST_DATA {
    set http_request_payload [HTTP::payload]
    set http_response_payload_encoded [b64encode $http_request_payload]
}

when HTTP_RESPONSE {
    set contentLength [HTTP::header "Content-Length"]
    set response_status_code [HTTP::status]
 
    set response_index 0
    foreach aHeader [HTTP::header names] {
        if {$response_index==0} {
            set hValue [HTTP::header value $aHeader]
            set responseHeaderValue "$aHeader:$hValue"
            set responseHeaderValueEncoded [b64encode $responseHeaderValue]
            set response_headers "\"$responseHeaderValueEncoded\""
        } else {
            set hValue [HTTP::header value $aHeader]
            set hValue [HTTP::header value $aHeader]
            set responseHeaderValue "$aHeader:$hValue"
            set responseHeaderValueEncoded [b64encode $responseHeaderValue]
            set response_headers "$response_headers,\"$responseHeaderValueEncoded\""
        }
        set response_index [expr {$response_index+1}]
    }
    set response_headers "\[$response_headers\]" 

 
    if {[HTTP::header exists "Content-Length"]}{ 
        HTTP::collect [HTTP::header "Content-Length"]
    } else {
        HSL::send $apiclarity_hsl "\{\"destination\":\"$destination_address\",\"requestID\":\"$request_id\",\"source\":\"$source_address\",\"scheme\":\"$scheme\",\"requestHost\":\"$request_host\",\"requestMethod\":\"$request_method\",\"requestPath\":\"$request_path\",\"responseStatus\":\"$response_status_code\",\"requestheaders\":$request_headers,\"requestpayload\":\"$http_request_payload_encoded\",\"responseheaders\":$response_headers\}\n"

    }
}

when HTTP_RESPONSE_DATA {
    set http_response_payload [HTTP::payload]
    set http_response_payload_encoded [b64encode $http_response_payload]

    HSL::send $apiclarity_hsl "\{\"destination\":\"$destination_address\",\"requestID\":\"$request_id\",\"source\":\"$source_address\",\"scheme\":\"$scheme\",\"requestHost\":\"$request_host\",\"requestMethod\":\"$request_method\",\"requestPath\":\"$request_path\",\"responseStatus\":\"$response_status_code\",\"requestheaders\":$request_headers,\"requestpayload\":\"$http_request_payload_encoded\",\"responseheaders\":$response_headers,\"responsepayload\":\"$http_response_payload_encoded\"\}\n"
}

when CLIENT_CLOSED {
    unset apiclarity_hsl
}