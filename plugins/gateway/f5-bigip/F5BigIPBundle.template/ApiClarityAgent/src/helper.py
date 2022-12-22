import traceback

import yaml
import logging
import os
import copy


def create_file(file: str):
    try:
        folder = os.path.dirname(file)
        os.makedirs(folder, exist_ok=True)
        open(file, "a").close()
    except Exception as e:
        raise Exception(f"Failed creating file {file}: {e}")


def __merge_config(config: dict, overwrite: dict):
    """
    overwrite values if present values overwrite config values
    :param dict1:
    :param dict2:
    :return:
    """
    for k in overwrite:
        if isinstance(overwrite[k], dict):
            if k in config:
                __merge_config(config[k], overwrite[k])
                continue

        config[k] = overwrite[k]


def __config_obfuscate(config: dict) -> dict:
    c = copy.deepcopy(config)
    c['apiclarity-token'] = '**************'
    return c


def get_configs(default_config_path="./config.yaml") -> dict:
    # Load defaults
    with open(default_config_path, 'r') as f:
        configs = yaml.safe_load(f) or {}

    config_filename = os.environ.get('CONFIG_PATH', None)
    if config_filename:
        try:
            with open(config_filename, 'r') as f:
                custom_configs = yaml.safe_load(f) or {}
                __merge_config(configs, custom_configs)
        except Exception as e:
            logging.error(f"Unable to open config file {config_filename}: {e}")

    logging.basicConfig(level=logging.DEBUG if configs['debug'] else logging.INFO,
                        format='%(asctime)s %(levelname)s [%(threadName)s] %(message)s')

    logging.debug("DEBUG IS ON")

    fatal = False
    if not configs.get('apiclarity-url'):
        logging.error("Missing config `apiclarity-url`")
        fatal = True
    if not configs.get('apiclarity-token'):
        logging.error("Missing config `apiclarity-token`")
        fatal = True
    if configs.get('remote-log-proto') not in ['TCP', 'UDP']:
        logging.error("`remote-log-proto` must be one of the following values `TCP`, `UDP`")
        fatal = True
    if configs.get("apiclarity-url").endswith("/"):
        configs["apiclarity-url"] = configs["apiclarity-url"][:-1]

    if fatal:
        raise Exception("Invalid configuration")

    logging.info("\n##################################### CONFIG #####################################\n"
                 + yaml.dump(__config_obfuscate(configs)) +
                 "##################################################################################")
    return configs

def log_exception(msg):
    logging.error(msg)
    if logging.getLogger('').isEnabledFor(logging.DEBUG):
        traceback.print_exc()
