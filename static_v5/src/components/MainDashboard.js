import React from 'react';
import { render } from 'react-dom';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';

import NodeDashboard from 'components/NodeDashboard';
import VisibleUDFDashboard from 'containers/VisibleUDFDashboard';
import VisibleClusterDashboard from 'containers/VisibleClusterDashboard';
import Welcome from 'components/Welcome';
import { VIEW_TYPE } from 'classes/constants';

class MainDashboard extends React.Component {
  constructor(props) {
    super(props);

    this.changeView = this.changeView.bind(this);
  }

  changeView(toView) {
    const { selectedEntityPath, view } = this.props.currentView;
    if (toView === view)
      return;

    this.props.onSelectPath(selectedEntityPath, toView);
  }

  render() {
    const { currentView, isClusterConnected, clusterName } = this.props;
    const { entities, viewType, view } = currentView;
    const { clusterID, nodeHost, namespaceName, setName, udfName } = currentView;

    let dashboard;

    if (clusterID && !isClusterConnected)
      dashboard = <h4 style={{marginTop: 20}}> Please connect to {`"${clusterName}"`} to continue </h4>;
    else if (viewType === VIEW_TYPE.NODE)
      dashboard = <NodeDashboard clusterID={clusterID} nodeHost={nodeHost} />
    else if (viewType === VIEW_TYPE.UDF || viewType === VIEW_TYPE.UDF_OVERVIEW)
      dashboard = <VisibleUDFDashboard />
    else if (viewType === VIEW_TYPE.CLUSTER)
      dashboard = <VisibleClusterDashboard />
    else
      dashboard = <Welcome />;

    return (
      <div>
        {dashboard}
      </div>
      );
  }
}

MainDashboard.PropTypes = {
  // current view state
  currentView: PropTypes.object,
  // select a path
  // onSelectPath(path, view)
  onSelectPath: PropTypes.func,
  // if a cluster entity is selected 
  // is the cluster connected
  isClusterConnected: PropTypes.bool,
  // name of the cluster the dashboard is in
  clusterName: PropTypes.bool,
};

export default MainDashboard;
