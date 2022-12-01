# APIClarity Apigee-X plugin

This is a Apigee-X plugin that integrates with APIClarity.
For more information about Apigee-X please visit: https://cloud.google.com/apigee/docs


## Prerequisites

* APIClarity backend is running and exposed externally through a publicly reachable URL
* There is an Apigee-X proxy configured in your GCP account and such Apigee-X have connectivity to the above APIClarity

## Install Instructions

1. **Add Certificate to Apigee-X**

    1. In Apigee-x console, go to Admin->Environments->TLS Keystores and create a keystore named `Apiclarity`
    2. Create an Alias named `Apiclarity` and upload the Apiclarity public certificate
    3. In Admin->Environments->References, create a reference named `apiclarityRef` that refers the certificate named `Apiclarity` you previously created.

2. **Configure the shared flow bundle**
   1. Prepare the bundle:
       Run the following script by populating the environment variables with APIClarity URL and the APIClarity token generated for the Apigee-X proxy:
       ```shell
       
       APICLARITY_URL=https://1.2.3.4:443/ \
       APICLARITY_APIGEEX_TOKEN=xxxxxyyyyzzzzz== \
       ./preparebundle.sh
       ```
    2. Upload the bundle: 
       * In Apigee-x console go to Develop->Shared Flows and upload abundle 
       * Select to upload the file ApiClarityFlowBundle.zip generated in previous step
    3. Deploy the shared flow created by the bundle
    4. In Apigee-x console, go to Admin->Environments->Flow Hooks and associate the shared flow to the Post-proxy hook 
    
