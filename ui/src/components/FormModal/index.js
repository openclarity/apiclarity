import React, { Component } from 'react';
import Modal from 'components/Modal';
import ModalConfirmation from 'components/ModalConfirmation';

export default class FormModal extends Component {
    state = {
        formDirty: false,
        showEditConfirmation: false
    }

    handleFormClose = () => {
        const {formDirty} = this.state;

        if (!formDirty) {
            this.onFormClose();
            return;
        }

        this.setState({showEditConfirmation: true});
    }

    closeEditConfirmation = () => {
        this.setState({showEditConfirmation: false});
    }

    onFormClose = () => {
        this.props.onClose();
    }

    render() {
        const {showEditConfirmation} = this.state;
        const {allowClose, className, center, loading, formComponent: FormComponent, formProps} = this.props;

        return (
            <React.Fragment>
                <Modal allowClose={allowClose} onClose={this.handleFormClose} className={className} center={center} loading={loading}>
                    <FormComponent {...formProps} onFormCancel={this.handleFormClose} onDirtyChanage={formDirty => this.setState({formDirty})} />
                </Modal>
                {showEditConfirmation && <ModalConfirmation
                                             title="Unsaved changes"
                                             message="You have unsaved changes. Are you sure you want to leave this page?"
                                             confirmTitle="Ok"
                                             onCancle={this.closeEditConfirmation}
                                             onConfirm={() => {
                                                 this.closeEditConfirmation();
                                                 this.onFormClose();
                                             }}
                                         />
                }
            </React.Fragment>
        );
    }
}
