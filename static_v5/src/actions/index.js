export const REQUEST_CLUSTERS = 'REQUEST_CLUSTERS';
function requestClusters() {
  return {
    type: REQUEST_CLUSTERS
  };
}

export const RECEIVE_CLUSTERS = 'RECEIVE_CLUSTERS';
function receiveClusters(clusters) {
  return {
    type: RECEIVE_CLUSTERS,
    clusters
  };
}

export const SELECT_CLUSTER = 'SELECT_CLUSTER';
export const SELECT_NODE = 'SELECT_NODE';
export const SELECT_NAMESPACE = 'SELECT_NAMESPACE';
export const clusterEntitySelected = (entity) => {
  return {
    // TODO type, cluster, node, namespace based on entity
    type: SELECT_NODE,
    node: entity,
  };
};

export const SELECT_ENTITY_VIEW = 'SELECT_ENTITY_VIEW';
export const entityViewSelected = (entity, view) => {
  return {
    type: SELECT_ENTITY_VIEW,
    entity: entity,
    view: view
  };
}

export const DISPLAY_ADD_CLUSTER_CONNECTION = 'DISPLAY_ADD_CLUSTER_CONNECTION';
export const displayAddClusterConnection = (display) => {
  return {
    type: DISPLAY_ADD_CLUSTER_CONNECTION,
    display: display,
  };
}

export const ADD_CLUSTER_CONNECTION = 'ADD_CLUSTER_CONNECTION';
export const ADDING_CLUSTER_CONNECTION = 'ADDING_CLUSTER_CONNECTION';
export function addClusterConnection(connection) {
  return function(dispatch) {
    dispatch({
      type: ADDING_CLUSTER_CONNECTION
    });

    // TODO send request to server
    setTimeout(() => {
      dispatch({
        type: ADD_CLUSTER_CONNECTION,
        connection: connection,
      });
    }, 2000);
  }
}


export const ENTITY_NODE_COLLAPSED = 'ENTITY_NODE_COLLAPSED';
export const ENTITY_NODE_EXPANDED = 'ENTITY_NODE_EXPANDED';
export const expandEntityNode = (node) => {
  return {
    type: ENTITY_NODE_EXPANDED,
    node: node
  };
}
export const collapseEntityNode = (node) => {
  return {
    type: ENTITY_NODE_COLLAPSED,
    node: node
  };
}

export function fetchClusters() {
  const dummyData = [{
    label: 'Cluster ONE',
    children: [{
      label: 'Cluster THREE',
      children: []
    }, {
      label: 'Cluster FOUR',
      children: []
    }]
  }, {
    label: 'Cluster TWO',
    children: []
  }];

  return function(dispatch) {
    dispatch(requestClusters());

    setTimeout(() => {
      dispatch(receiveClusters(dummyData));
    }, 200);
  }
}


