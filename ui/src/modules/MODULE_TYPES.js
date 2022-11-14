import { ICON_NAMES } from 'components/Icon';
import COLORS from 'utils/scss_variables.module.scss';

const MODULE_TYPES = {
    EVENT_DETAILS: 'EVENT_DETAILS',
    INVENTORY_DETAILS: 'INVENTORY_DETAILS'
};


const MODULE_STATUS_TYPES_MAP = {
    ALERT_WARN: {
        value: "WARN",
        label: "Warning",
        icon: ICON_NAMES.ALERT,
        color: COLORS["color-risk-high"]
    },
    ALERT_INFO: {
        value: "INFO",
        label: "Information",
        icon: ICON_NAMES.INFO,
        color: COLORS["color-risk-unknown"]
    },
    ALERT_CRITICAL: {
        value: "CRITICAL",
        label: "Critical",
        icon: ICON_NAMES.ALERT,
        color: COLORS["color-risk-critical"]
    }
};

export { MODULE_TYPES, MODULE_STATUS_TYPES_MAP };
