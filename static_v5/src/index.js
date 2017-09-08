import React from 'react';
import { render } from 'react-dom';
import { createStore, applyMiddleware } from 'redux';
import { Provider } from 'react-redux';
import thunkMiddleware from 'redux-thunk';
import $ from 'jquery';

import app from './reducers';
import { fetchClusters } from './actions/clusters';
import VisibleApp from './containers/VisibleApp';
import reduxMiddleware from 'classes/reduxMiddleware';

// import all css
import 'bootstrap/dist/css/bootstrap.css';
import 'font-awesome/css/font-awesome.css';
import 'nvd3/build/nv.d3.css';
import 'react-widgets/lib/less/react-widgets.less';
import 'ag-grid/dist/styles/ag-grid.css';
import 'ag-grid/dist/styles/theme-material.css';

import './styles/index.scss';

// whatwg-fetch needs a Promise class to be available on the window object to
// work. To be compatible with older browsers need to add a polyfill.
import Promise from 'bluebird';
if (!window.Promise) {
  window.Promise = Promise;
}

// see http://jquense.github.io/react-widgets/docs/#/i18n?_k=gqx37t
import moment from 'moment';
import momentLocalizer from 'react-widgets/lib/localizers/moment';
momentLocalizer(moment);

const store = createStore(
  app,
  applyMiddleware(
    thunkMiddleware,
    reduxMiddleware
  )
);

// remove loader
$('#apploading').remove();

render(
  <Provider store={store}>
    <VisibleApp />
  </Provider>
  , document.getElementById('app'));
