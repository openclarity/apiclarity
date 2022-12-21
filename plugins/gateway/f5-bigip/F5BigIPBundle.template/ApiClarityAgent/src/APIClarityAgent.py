import json
import sys
import threading
import queue
import socket
import time
import requests
import helper
import ssl
from APIClarityHelper import prepare_telemetry
import logging
from urllib.parse import urlparse
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

QUEUE_MAX_SIZE = 100
__configs = None


def trace_receiver(trace_queue: queue.Queue):
    """
    Receive traces from network and put them in a queue to be processed.
    Traces a separated by new line character.
    If queue is full, trace is dropped on the floor.
    """
    port = __configs['remote-log-port']
    proto = __configs['remote-log-proto']

    logging.info(f"Starting server on {proto}://0.0.0.0:{port}")

    try:
        ssocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM if proto == "TCP" else socket.SOCK_DGRAM)
        ssocket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        ssocket.bind(('', port))

        if proto == "TCP":
            ssocket.listen()
    except Exception as e:
        helper.log_exception(f"Unable to start server: {e}")
        sys.exit(1)

    while True:
        try:
            if proto == "TCP":
                s, client = ssocket.accept()
                logging.info(f"Connection accepted from {client}")
            else:
                s = ssocket
            f = s.makefile()
        except Exception as e:
            helper.log_exception(f"Unable to accept new connection: {e}")
            continue

        while True:
            try:
                trace = f.readline()
                if not trace:
                    logging.info(f"Connection closed from client {client}")
                    break
                trace = trace.rstrip()
                logging.debug(f"Received a trace: {trace}")
                try:
                    jtrace = json.loads(trace)
                except Exception as e:
                    helper.log_exception(f"Unable to parse trace {trace}: {e}")
                    continue
                if logging.getLogger('').isEnabledFor(logging.DEBUG):
                    logging.debug("trace: " + json.dumps(jtrace, indent=2))
                if trace_queue.qsize() == trace_queue.maxsize:
                    logging.debug("Dropping trace")
                    continue
                logging.debug("Forwarding trace")
                trace_queue.put(jtrace)
            except Exception as e:
                helper.log_exception(f"Unable to handle trace {trace}: {e}")


def extract_destination_host(trace) -> str:
    return trace["requestHost"]



def create_apiclarity_session() -> (requests.Session, dict):
    # Verify certificate
    context = ssl.create_default_context()
    context.load_verify_locations(__configs['apiclarity-cert-path'])
    apiclarity_url = urlparse(__configs['apiclarity-url'])
    with socket.create_connection((apiclarity_url.hostname, apiclarity_url.port)) as sock:
        with context.wrap_socket(sock, server_hostname=__configs['apiclarity-cert-hostname']):
            logging.info(f"APIClarity certificate from {__configs['apiclarity-url']}  succesfully verified against hostname `__configs['apiclarity-cert-hostname']`")

    sess = requests.Session()
    token = __configs['apiclarity-token']

    sess.verify = False #__configs['apiclarity-cert-path']
    # sess.verify = __configs['apiclarity-cert-path']
    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json",
        "X-Trace-Source-Token": token.encode("UTF-8")
    }
    return sess, headers


def trace_sender(trace_queue: queue.Queue, new_api_queue: queue.Queue, api_inventory: dict):
    logging.info("started")

    sess, headers = create_apiclarity_session()
    telemetry_path = __configs['apiclarity-url'] + "/api/telemetry"
    while True:
        trace = trace_queue.get()
        logging.debug(f"Received trace from queue: {trace['requestID']}")
        try:
            dhost = extract_destination_host(trace)
        except Exception as e:
            logging.debug(f"Unable to extract destinaton host from trace {trace['requestID']}: {e}")
            continue

        if dhost not in api_inventory['discovered'] and dhost not in api_inventory['to-notify']:
            logging.info(f"New API discovered: {dhost}")
            api_inventory['to-notify'].add(dhost)
            new_api_queue.put(dhost)
        if dhost not in api_inventory['to-trace']:
            logging.debug(f"Dropping trace not to trace: {trace['requestID']}")
            continue
        try:
            # send trace
            telemetry = prepare_telemetry(trace)
            response = sess.post(telemetry_path, json=telemetry, headers=headers)
            response.raise_for_status()
            if logging.getLogger('').isEnabledFor(logging.DEBUG):
                logging.debug("trace: " + json.dumps(telemetry, indent=2))
        except Exception as e:
            helper.log_exception(f"Unable to send telemetry {telemetry}: {e}")

    sess.close()


def api_notifier(new_api_queue: queue.Queue, api_inventory: dict):
    logging.info("started")

    sess, headers = create_apiclarity_session()
    discovered_apis_path = __configs['apiclarity-url'] + "/api/control/newDiscoveredAPIs"
    while True:
        notified = False
        newapi = new_api_queue.get()
        logging.info(f"notifying about new api {newapi}")
        try:
            newapis = {
                "hosts": [newapi]
            }
            response = sess.post(discovered_apis_path, json=newapis, headers=headers)
            if response.status_code < 200 or response.status_code >= 300:
                logging.error(
                    f"Unable to notify about discovered api {newapi}: return code={response.status_code}")
            else:
                notified = True
        except Exception as e:
            helper.log_exception(f"Unable to notify about discovered api {newapi}: {e}")
            api_inventory['to-notify'].remove(newapi)
            continue

        if notified:
            logging.info(f"Successfully notified about new api {newapi}")
            api_inventory['discovered'].add(newapi)


def hosts_to_trace_poller(api_inventory: dict):
    logging.info("started")

    sess, headers = create_apiclarity_session()
    hosts_to_trace_path = __configs['apiclarity-url'] + "/api/hostsToTrace"
    while True:
        try:
            logging.debug("Retrieving hosts to trace")
            response = sess.get(hosts_to_trace_path, headers=headers)
            response.raise_for_status()
            hosts_to_trace = response.json()["hosts"]
            logging.info(f"Retrieved hosts to trace: {hosts_to_trace}")
            api_inventory['to-trace'] = set(hosts_to_trace)
        except Exception as e:
            helper.log_exception(f"Unable to retrieve hosts to trace: {e}")
        time.sleep(int(__configs['refresh-interval-seconds']))


def main():
    global __configs
    __configs = helper.get_configs()
    # Queue for traces
    trace_queue = queue.Queue(maxsize=QUEUE_MAX_SIZE)

    # Queue for discovered APIs
    new_api_queue = queue.Queue()

    # API Inventory
    api_inventory = {
        'to-notify': set(),
        'discovered': set(),
        'to-trace': set()
    }

    t_receiver = threading.Thread(target=trace_receiver, name="trace-receiver", args=[trace_queue]).start()
    t_sender = threading.Thread(target=trace_sender, name="trace-sender",
                                args=[trace_queue, new_api_queue, api_inventory]).start()
    a_notifier = threading.Thread(target=api_notifier, name="api-notifier", args=[new_api_queue, api_inventory]).start()
    htt_poller = threading.Thread(target=hosts_to_trace_poller, name="hosts-to-trace-poller",
                                  args=[api_inventory]).start()


if __name__ == '__main__':
    main()
