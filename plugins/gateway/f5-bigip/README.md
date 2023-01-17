# APIClarity F5 BIG-IP LTM plugin

This is an F5 BIG-IP LTM plugin that integrates with APIClarity.
For more information about F5 BIG-IP LTM please visit: https://www.f5.com/products/big-ip-services/local-traffic-manager


## Prerequisites

* APIClarity backend is running and exposed externally through a publicly reachable URL
* F5 BIG-IP is running
* Another host (e.g. a Virtual Machine) able to run APIClarity Agent as docker container is running. This host must be reachable by F5 BIG-IP and must be able to reach APIclarity URL. Let us refer to this host as the `F5-Log-Collector`.

## Install Instructions
1. **Prepare the bundle**

 Run the following script by populating the environment variables with APIClarity URL, the APIClarity token and the path to the APIClarity certificate generated for the F5 BIG-IP instance:

       ```
       APICLARITY_URL=https://1.2.3.4:443/ \
       APICLARITY_TOKEN=xxxxxyyyyzzzzz== \
       APICLARITY_CERT_PATH=/path/to/apiclarity.crt \
       ./preparebundle.sh
       ```
 The generated bundle is named `F5BigIPBundle` and contains all the files needed for subsequent steps.

2. **Configure High Speed Logging in BIG-IP**

    1. Connect to BIG-IP management console
    2. In `System`->`Configuration`->`Log Destinations` click on the `+` button
    3. Fill the form as follows:
       * Name: `ApiClarityAgent`
       * Description: `APICLarity Agent receiving hsl`
       * Type: `Management Port`
       * Address: [The host where F5-Log-Collector is reachable from BIG-IP]
       * Port: `10514`
       * Protocol: `udp`
    5. In `System`->`Configuration`->`Log Publishers` click on the `+` button
    6. Fill the form as follows:
       * Name: `APIClarityLogPublisher`
       * Description: `ApiClarity Log Publisher`
       * Destinations: Make sure `ApiClarityAgent` is in the list of selected destinations 

3. **Configure Log Publisher iRule in BIG-IP**

    1. Connect to BIG-IP management console
    2. In `Local Traffic`->`iRules`->`iRule List` click on the `+` button
    3. Fill the form as follows:
       * Name: `APIClarityLogPublisher`
       * Definition: From the bundle folder, copy the content of `F5BigIPBundle\iRule\APIClarityLogPublisher.tcl`

4. **Deploy APIClarity Agent on the F5-Log-Collector**

   1. From the bundle copy the folder `F5BigIPBundle/ApiClarityAgent` to F5-Log-Collector host
   2. Connect to F5-Log-Collector host
   3. Build the ApiClarityAgent container image:
      * Go in the folder `ApiClarityAgent`
      * Execute: 
        ```
        dockerbuild.sh
        ```
   4. Launch the ApiClarityAgent:
      * Go in the folder `ApiClarityAgent/deploy/docker`
      * Execute:
        ```
        launch.sh
        ```
