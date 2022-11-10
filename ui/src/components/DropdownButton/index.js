import React, { Component } from 'react';
import { Dropdown, DropdownToggle, DropdownMenu } from 'reactstrap';
import classnames from 'classnames';

import './dropdown-button.scss';

class DropdownButtonDefaultOpen extends Component {
    state = {
        isOpen: false
    }

    toggleOpen = (isOpen) => {
        this.setState({isOpen});
    }

    render() {
        const {className, toggleButton, children, withCaret, disabled} = this.props;
        const {isOpen} = this.state;

        return (
            <Dropdown
                className={classnames("ps-dropdown", {[className]: className}, {disabled: disabled})}
                isOpen={isOpen}
                toggle={() => !disabled && this.toggleOpen(!isOpen)}
                direction="down"
            >
                <DropdownToggle tag="div" caret={withCaret}>{toggleButton}</DropdownToggle>
                <DropdownMenu>{children}</DropdownMenu>
            </Dropdown>
        );
    }
}

class DropdownButtonManualControl extends Component {
    toggleOpen = (isOpen) => {
        this.props.onToggle(isOpen);
    }

    render() {
        const {className, toggleButton, children, withCaret, isOpen, disabled} = this.props;

        return (
            <Dropdown
                className={classnames("ps-dropdown", {[className]: className}, {disabled: disabled})}
                isOpen={isOpen}
                toggle={() => {}}
                direction="down"
            >
                <DropdownToggle tag="div" caret={withCaret} onClick={() => !disabled && this.toggleOpen(!isOpen)}>{toggleButton}</DropdownToggle>
                <DropdownMenu>{children}</DropdownMenu>
            </Dropdown>
        );
    }
}

const DropdownButton = ({manualOpen, ...props}) => {
    if (manualOpen) {
        return <DropdownButtonManualControl {...props} />;
    }

    return <DropdownButtonDefaultOpen {...props} />;
}

export default DropdownButton;
