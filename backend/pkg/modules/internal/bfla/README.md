# Broken Function Level Authorization (BFLA)

Detection of potential BFLA issues for internal services in a Kubernetes cluster

This module will observe traffic in real time and build a model of service interactions.
We call it authorization model, because this in essence represents which service interactions are allowed and which are not allowed.

The process consists of two phases
 1. Learning phase
 2. Detection phase

## Learning
The learning phase begins upon user request or it can be configured to start as soon as a user uploads or reconstructs a spec for a given API.
By default, the learning lasts for 100 API calls (for an API), the learning can be stopped, reset or prolonged for a longer period.

![img.png](../../assets/bfla/images/img_1.png)

**(Orders) -> POST /v1/checkout -> (Payment)**: This call is registered during the learning phase.

## Detection
Once the learning phase has ended, the user can ask to start the detection phase.
During the detection phase, all the API calls that do not comply with the authorization model will be marked as warnings.
If the call was rejected by the API the warning will have a lower severity, however if the API accepted the unexpected call and returned a 2xx status code it will have a higher severity (check fig1 and fig2).

![img_2.png](../../assets/bfla/images/img_2.png)

**(Booking) -> POST /v1/checkout (2xx) -> (Payment)**: This call was not registered in the authorization model during the learning phase so the BFLA module will raise a warning. And since the response has a 2xx status code the warning has a high severity.

![img_3.png](../../assets/bfla/images/img_3.png)

**(Booking) -> POST /v1/checkout (4xx) -> (Payment)**: This call was not registered in the authorization model during the learning phase so the BFLA module will raise a warning. The warning has a lower severity because the API rejected this call.

## Module interaction

In case of a false positive, meaning that a certain service interaction was wrongly marked as legitimate, 
the user can mark it as illegitimate and if service interaction was wrongly marked as illegitimate, the user can mark it as legitimate.

The user can also stop, start, and reset learning.

List of all operations for tuning the authorization model:
 1. Mark event as legitimate.  
 2. Mark event as illegitimate.
 3. Stop learning.
 4. Start learning.
 5. Reset learning.

## Principal detection
Principal is the actor that makes the API call.
We detect the principal by matching known authorization protocols on the event headers.

Supported protocols:
1. Basic auth: The principal ID is the username from the `base64(username:password)` formula.
2. JWT: The principal ID is the Subject claim in the body.
3. X-Customer-ID header: The Principal ID is given by the Kong gateway when using authorization plugins.

## BFLA State Machine  
The following diagram depicts the BFLA State Machine. 
![img.png](../../assets/bfla/images/BFLA_FSM.png)

For reference we also include the .dot state machine definition: [BFLA_FSM.dot](../../assets/bfla/BFLA_FSM.dot)
