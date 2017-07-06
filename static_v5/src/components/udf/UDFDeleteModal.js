import React from 'react';
import { render } from 'react-dom';
import PropTypes from 'prop-types';

import { deleteUDF } from 'api/udf';
import Spinner from 'components/Spinner';

import { Button, Modal, ModalHeader, ModalBody, ModalFooter } from 'reactstrap';

class UDFDeleteModal extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      inProgress: false,
      successful: false,
      errorMsg: '',
    };

    this.onDeleteUDF = this.onDeleteUDF.bind(this);
    this.onDeleteSuccess = this.onDeleteSuccess.bind(this);
    this.onCancel = this.onCancel.bind(this);
  }

  onDeleteSuccess() {
    this.props.onDeleteSuccess();
  }

  onCancel() {
    this.props.onCancel();
  }

  onDeleteUDF() {
    const { clusterID, udfName } = this.props;
    this.setState({
      inProgress: true
    });
    deleteUDF(clusterID, udfName)
      .then(() => {
        this.setState({
          inProgress: false,
          successful: true
        });
        window.setTimeout(() => this.onDeleteSuccess(), 2000);
      })
      .catch((msg) => {
        this.setState({
          inProgress: false,
          successful: false,
          errorMsg: msg || 'Failed to delete UDF'
        });
      });
  }

  render() {
    const { inProgress, successful, errorMsg } = this.state;

    const disabled = inProgress || successful;
    if (!inProgress && successful) {
      return (
        <Modal isOpen={true} toggle={() => {}}>
          <ModalHeader> Success </ModalHeader>
          <ModalBody> Successfully deleted {this.props.udfName} </ModalBody>
        </Modal>
      );
    }

    return (
      <Modal isOpen={true} toggle={() => {}}>
        <ModalHeader> Confirm </ModalHeader>
        <ModalBody>  Delete {this.props.udfName} ?  </ModalBody>
        <ModalFooter>
          {!inProgress && successful &&
            errorMsg}
          {inProgress &&
           <span> <Spinner /> Deleting ... </span>}
          <Button disabled={disabled} color="primary" onClick={this.onDeleteUDF}>Confirm</Button>
          <Button disabled={disabled} color="secondary" onClick={this.onCancel}>Cancel</Button>
        </ModalFooter>
      </Modal>
    );
  }
}

UDFDeleteModal.PropTypes = {
  clusterID: PropTypes.string,
  udfName: PropTypes.string,
  // callback on successfull delete
  // onDeleteSuccess()
  onDeleteSuccess: PropTypes.func,
  // callback on cancel
  // onCancel()
  onCancel: PropTypes.func,
};

export default UDFDeleteModal;

