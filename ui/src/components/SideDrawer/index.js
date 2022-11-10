import React, { Component } from 'react';
import ReactDOM from 'react-dom';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import Loader from 'components/Loader';
import CloseButton from 'components/CloseButton';
import { ModalSection, ModalTitleDataItem } from './utils';

import './side-drawer.scss';

export {
    ModalSection,
    ModalTitleDataItem
}

class Modal extends Component {
    state = {
        container: null
    }

    componentDidMount() {
        const container = document.querySelector("main[role='main']");
        if (!!container && container !== this.state.container) {
            this.setState({container});
        }
    }

    render() {
        const {container} = this.state;

        if (!container) {
            return null;
        }

        return ReactDOM.createPortal(
            <ModalInner {...this.props} />,
            container
        );
    }
}

const ModalInner = ({children, onClose, className, allowClose=true, loading=false, center=false, centerLarge=false}) => {
    const onOuterClick = event => {
        event.stopPropagation();
        event.preventDefault();

        if (allowClose) {
            onClose();
        }
    }

    return (
        <div className={classnames("modal-outer", {center}, {"center-large": centerLarge})} onClick={onOuterClick}>
            <div className={classnames("modal-container", {[className]: !isEmpty(className)})} onClick={(event) => event.stopPropagation()}>
                {allowClose && <CloseButton onClose={onClose} />}
                {loading ? <Loader /> : children}
            </div>
        </div>
    );
}

export default Modal;
