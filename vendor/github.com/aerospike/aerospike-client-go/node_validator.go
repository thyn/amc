// Copyright 2013-2017 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"net"
	"strings"
	"sync"

	. "github.com/aerospike/aerospike-client-go/logger"
	. "github.com/aerospike/aerospike-client-go/types"
)

type nodesToAddT struct {
	nodesToAdd map[string]*Node
	mutex      sync.RWMutex
}

func (nta *nodesToAddT) addNodeIfNotExists(ndv *nodeValidator, cluster *Cluster) bool {
	nta.mutex.Lock()
	defer nta.mutex.Unlock()

	_, exists := nta.nodesToAdd[ndv.name]
	if !exists {
		// found a new node
		node := cluster.createNode(ndv)
		nta.nodesToAdd[ndv.name] = node
	}
	return exists
}

// Validates a Database server node
type nodeValidator struct {
	name        string
	aliases     []*Host
	primaryHost *Host

	conn *Connection

	supportsFloat, supportsBatchIndex, supportsReplicasAll, supportsGeo, supportsPeers bool
}

func (ndv *nodeValidator) seedNodes(cluster *Cluster, host *Host, nodesToAdd *nodesToAddT) error {
	if err := ndv.setAliases(host); err != nil {
		return err
	}

	found := false
	var resultErr error
	for _, alias := range ndv.aliases {
		if resultErr = ndv.validateAlias(cluster, alias); resultErr != nil {
			Logger.Debug("Alias %s failed: %s", alias, resultErr)
			continue
		}

		found = true
		if !nodesToAdd.addNodeIfNotExists(ndv, cluster) {
			ndv.conn.Close()
		}
	}

	if !found {
		return resultErr
	}
	return nil
}

func (ndv *nodeValidator) validateNode(cluster *Cluster, host *Host) error {
	if err := ndv.setAliases(host); err != nil {
		return err
	}

	var resultErr error
	for _, alias := range ndv.aliases {
		if err := ndv.validateAlias(cluster, alias); err != nil {
			resultErr = err
			Logger.Debug("Aliases %s failed: %s", alias, err)
			continue
		}
		return nil
	}

	return resultErr
}

func (ndv *nodeValidator) setAliases(host *Host) error {
	// IP addresses do not need a lookup
	ip := net.ParseIP(host.Name)
	if ip != nil {
		aliases := make([]*Host, 1)
		aliases[0] = NewHost(host.Name, host.Port)
		ndv.aliases = aliases
	} else {
		addresses, err := net.LookupHost(host.Name)
		if err != nil {
			Logger.Error("HostLookup failed with error: ", err)
			return err
		}
		aliases := make([]*Host, len(addresses))
		for idx, addr := range addresses {
			aliases[idx] = NewHost(addr, host.Port)
		}
		ndv.aliases = aliases
	}
	Logger.Debug("Node Validator has %d nodes and they are: %v", len(ndv.aliases), ndv.aliases)
	return nil
}

func (ndv *nodeValidator) validateAlias(cluster *Cluster, alias *Host) error {
	conn, err := NewSecureConnection(&cluster.clientPolicy, alias)
	if err != nil {
		return err
	}

	// need to authenticate
	if err := conn.Authenticate(cluster.user, cluster.Password()); err != nil {
		// Socket not authenticated. Do not put back into pool.
		conn.Close()

		return err
	}

	// check to make sure we have actually connected
	info, err := RequestInfo(conn, "build")
	if err != nil {
		// Socket not authenticated. Do not put back into pool.
		conn.Close()

		return err
	}
	if _, exists := info["ERROR:80:not authenticated"]; exists {
		// Socket not authenticated. Do not put back into pool.
		conn.Close()

		return NewAerospikeError(NOT_AUTHENTICATED)
	}

	hasClusterName := len(cluster.clientPolicy.ClusterName) > 0

	var infoKeys []string
	if hasClusterName {
		infoKeys = []string{"node", "features", "cluster-id"}
	} else {

		infoKeys = []string{"node", "features"}
	}
	infoMap, err := RequestInfo(conn, infoKeys...)
	if err != nil {
		return err
	}

	nodeName, exists := infoMap["node"]
	if !exists {
		return NewAerospikeError(INVALID_NODE_ERROR)
	}

	// set features
	if features, exists := infoMap["features"]; exists {
		ndv.setFeatures(features)
	}

	ndv.name = nodeName
	ndv.primaryHost = alias
	ndv.conn = conn

	return nil
}

func (ndv *nodeValidator) setFeatures(features string) {
	featureList := strings.Split(features, ";")
	for i := range featureList {
		switch featureList[i] {
		case "float":
			ndv.supportsFloat = true
		case "batch-index":
			ndv.supportsBatchIndex = true
		case "replicas-all":
			ndv.supportsReplicasAll = true
		case "geo":
			ndv.supportsGeo = true
		case "peers":
			ndv.supportsPeers = true
		}
	}
}
