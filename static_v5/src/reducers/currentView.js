import { SELECT_NODE_VIEW, SELECT_CLUSTER_VIEW, INITIALIZE_VIEW } from 'actions/currentView';
import { SELECT_START_VIEW, SELECT_NAMESPACE_VIEW, SELECT_SET_VIEW } from 'actions/currentView';
import { SELECT_SET_OVERVIEW } from 'actions/currentView';
import { SELECT_UDF_VIEW, SELECT_UDF_OVERVIEW, SHOW_LEFT_PANE, HIDE_LEFT_PANE } from 'actions/currentView';
import { SELECT_INDEX, SELECT_INDEXES_OVERVIEW, SELECT_CLUSTER_ON_STARTUP } from 'actions/currentView';
import { SELECT_VIEW, SELECT_VIEW_FOR_VIEW_TYPE } from 'actions/currentView';
import { SELECT_LOGICAL_NAMESPACE, SELECT_LOGICAL_CLUSTER } from 'actions/currentView';
import { SELECT_LOGICAL_VIEW, SELECT_PHYSICAL_VIEW } from 'actions/currentView';
import { VIEW_TYPE } from 'classes/constants';
import { updateURL } from 'classes/urlAndViewSynchronizer';
import { LOGICAL_CLUSTER_ACTIONS, CLUSTER_ACTIONS } from 'classes/entityActions';

// store the previous physical, logical state between view changes
let PhysicalState = {};
let LogicalState = {};

function isNotSet(obj) {
  return obj === undefined || obj === null || Object.keys(obj).length === 0;
}

function getViewState(state) {
  const props = ['viewType', 'view', 'clusterID', 'nodeHost', 'namespaceName', 
                 'setName', 'udfName', 'indexName'];

  let obj = {};
  props.forEach((v) => {
    obj[v] = state[v];
  });

  return obj;
}

const InitState = {
  // is the current view initialized
  isInitialized: false,

  // type of view
  // see VIEW_TYPE
  viewType: null,
  // the view of interest of the VIEW_TYPE
  // like Performance, Machine, Storage
  view: null, 

  // whether the left pane holding the entity tree is shown
  showLeftPane: true,

  // the entities in the selected view
  // ex: namespace is identified by clusterID, nodeHost, namespaceName
  clusterID: null,
  nodeHost: null,
  namespaceName: null,
  setName:null,
  udfName: null,
  indexName: null,
};

function updateEntityView(state) {
  return (update) => {
    const NullEntities = {
      clusterID: null,
      nodeHost: null,
      namespaceName: null,
      setName:null,
      udfName: null,
      indexName: null,
    };

    return Object.assign({}, state, NullEntities, update);
  };
}

// the current state of the view of the app
export default function currentView(state = InitState, action) {
  const updateFn = updateEntityView(state);
  const wasInitialized = state.isInitialized;
  let updated, w;

  switch (action.type) {
    case SELECT_PHYSICAL_VIEW:
      if (isNotSet(PhysicalState) && state.clusterID) {
        // select physical overview of corresponding cluster
        w = {
          viewType: VIEW_TYPE.CLUSTER,
          view: CLUSTER_ACTIONS.Overview,
          clusterID: state.clusterID,
        };
      } else {
        // restore previous physical view
        w = Object.assign({}, {
          viewType: VIEW_TYPE.START_VIEW,
          view: null,
        }, PhysicalState);
      }
      updated = updateFn(w);

      // save current logical view
      LogicalState = getViewState(state);
      break;

    case SELECT_LOGICAL_VIEW:
      if (isNotSet(LogicalState) && state.clusterID) {
        // select logical overview of corresponding cluster
        w = {
          viewType: VIEW_TYPE.LOGICAL_CLUSTER,
          view: LOGICAL_CLUSTER_ACTIONS.Overview,
          clusterID: state.clusterID,
        };
      } else { // restore previous logical view
        w = Object.assign({}, {
          viewType: VIEW_TYPE.LOGICAL_START_VIEW,
          view: null,
        }, LogicalState);
      }
      updated = updateFn(w);

      // save current physical view
      PhysicalState = getViewState(state);
      break;

    case SELECT_LOGICAL_NAMESPACE:
      updated = updateFn({
        clusterID: action.clusterID,
        namespaceName: action.namespaceName,
        view: action.view,

        viewType: VIEW_TYPE.LOGICAL_NAMESPACE,
      });
      break;

    case SELECT_LOGICAL_CLUSTER:
      updated = updateFn({
        clusterID: action.clusterID,
        view: action.view,

        viewType: VIEW_TYPE.LOGICAL_CLUSTER,
      });
      break;

    case SELECT_VIEW:
      const nv = action.newView;
      updated = Object.assign({}, state, {
        view:          nv.view,
        viewType:      nv.viewType,

        clusterID:     nv.clusterID || null,
        nodeHost:      nv.nodeHost || null,
        namespaceName: nv.namespaceName || null,
        setName:       nv.setName || null,
        udfName:       nv.udfName || null,
        indexName:     nv.indexName || null,
      });
      break;

    case SELECT_VIEW_FOR_VIEW_TYPE:
      updated = Object.assign({}, state, {
        view: action.view
      });
      break;

    case SELECT_CLUSTER_VIEW:
      updated = updateFn({
        view: action.view,

        viewType: VIEW_TYPE.CLUSTER,
        clusterID: action.clusterID
      });
      break;

    case SELECT_CLUSTER_ON_STARTUP:
      updated = state;

      const v = state.viewType;
      if (v === null || v === VIEW_TYPE.START_VIEW) {
        updated = updateFn({
          view: action.view,

          viewType: VIEW_TYPE.CLUSTER,
          clusterID: action.clusterID
        });
      }
      break;

    case SELECT_NODE_VIEW:
      updated =  updateFn({
        view: action.view,

        viewType: VIEW_TYPE.NODE,
        clusterID: action.clusterID,
        nodeHost: action.nodeHost,
      });
      break;

    case SELECT_NAMESPACE_VIEW:
      updated = updateFn({
        view: action.view,

        viewType: VIEW_TYPE.NAMESPACE,
        clusterID: action.clusterID,
        nodeHost: action.nodeHost,
        namespaceName: action.namespaceName,
      });
      break;

    case SELECT_SET_VIEW:
      updated =  updateFn({
        view: action.view,

        viewType: VIEW_TYPE.SET,
        clusterID: action.clusterID,
        nodeHost: action.nodeHost,
        namespaceName: action.namespaceName,
        setName: action.setName,
      });
      break;

    case SELECT_SET_OVERVIEW:
      updated =  updateFn({
        view: action.view,

        viewType: VIEW_TYPE.SET_OVERVIEW,
        clusterID: action.clusterID,
        nodeHost: action.nodeHost,
        namespaceName: action.namespaceName,
      });
      break;

    case SELECT_UDF_VIEW:
      updated =  updateFn({
        view: action.view,

        viewType: action.viewType,
        clusterID: action.clusterID,
        udfName: action.udfName,
      });
      break;


    case SELECT_UDF_OVERVIEW:
      updated =  updateFn({
        view: action.view, 

        viewType: action.viewType,
        clusterID: action.clusterID,
      });
      break;

    case SELECT_INDEXES_OVERVIEW:
      updated =  updateFn({
        view: action.view,
        
        viewType: action.viewType,
        clusterID: action.clusterID,
      });
      break;

    case SELECT_INDEX:
      updated =  updateFn({
        view: action.view,

        viewType: action.viewType,
        clusterID: action.clusterID,
        indexName: action.indexName
      });
      break;

    case INITIALIZE_VIEW:
      updated = Object.assign({}, state, {
        isInitialized: true
      });
      break;

    case SELECT_START_VIEW:
      updated = updateFn({
        viewType: VIEW_TYPE.START_VIEW,
      });
      break;

    case HIDE_LEFT_PANE:
      updated = Object.assign({}, state, {
        showLeftPane: false
      });
      break;

    case SHOW_LEFT_PANE:
      updated = Object.assign({}, state, {
        showLeftPane: true
      });
      break;

    default:
      updated = state;
      break;
  }

  // update the URL to reflect the view
  if (wasInitialized)
    updateURL(updated);
  
  return updated;
}
