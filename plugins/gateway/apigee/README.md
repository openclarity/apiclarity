# Apigee policy and telemetry visualization
 
## Description
This section describes the steps used in setting up apigee proxy, creating a policy and visualizing the end results in apiclarity.
 
 
## Starting condition
* A running deployment of APIClarity with a service to expose it through LoadBalancer so as to publish traffic from Apigee policy, also make a note of the LB IP address.
 The APIClarity telemetry url should be something like <API_CLARITY_SERVICE:9000/api/telemetry>. This service endpoint will be used in the javascript policy.
 N/B: Make sure not to change the existing apicalrity service cluster type from ClusetrIP to LoadBalancer, rather make a copy of the original apiclarity and define cluster type as Load Balancer
* A running k8s cluster where backend service to be exposed from outside will reside.
* URL of backend service (BACKEND_SERVICE_URL) is noted.
* An Apigee instance is currently provisioned with endpoint to gateway (APIGEE_GATEWAY_URL) noted
 
## Instructions
* Create a simple passthrough proxy which exposes the backend service (with url BACKEND_SERVICE_URL) on apigee
* Add the BACKEND_SERVICE_URL of the backend service as target endpoint on the passthrough proxy created above.
* Click on "Developer" Tab on Apigee UI
* Select "PreFlow" under Target endpoints -> click the Step+ button under response area
* Scroll down the popup window and select Javascript as policy
* Enter policy name (or leave all as default) and click on Add
* Past the content of the "github.com/apiclarity/plugins/gateway/apigee/telemetry-policy.js" into the newly created javascript policy
* Replace <API_CLARITY_SERVICE:9000/api/telemetry> with noted above
 
## Debugging
* Start a debugging session by clicking on the "Debug" Tap.
* Select the latest entry (e.g. eval-xxx) in the Env. drop down menu and click on the Start Debug Session
* Tail the logs of the apiclarity pod, after setting debugging to true on the apiclarity deployment.yaml