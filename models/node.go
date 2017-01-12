package models

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	as "github.com/aerospike/aerospike-client-go"

	"github.com/citrusleaf/amc/common"
	"github.com/citrusleaf/amc/rrd"
)

type NodeStatus string
type NodeVisibilityStatus string
type XDRStatus string

var nodeStatus = struct {
	On, Off NodeStatus
}{
	"on", "off",
}

var xdrStatus = struct {
	On, Off XDRStatus
}{
	"on", "off",
}

var nodeVisibilityStatus = struct {
	On, Off NodeVisibilityStatus
}{
	"on", "off",
}

type Node struct {
	cluster  *Cluster
	origNode *as.Node
	origHost *as.Host

	status  NodeStatus
	visible NodeVisibilityStatus

	namespaces map[string]*Namespace

	latestInfo, oldInfo common.Info
	latestConfig        common.Stats
	latestNodeLatency   map[string]common.Stats

	stats, nsAggStats common.Stats
	nsAggCalcStats    common.Stats

	statsHistory   map[string]*rrd.Bucket
	latencyHistory *rrd.SimpleBucket

	serverTime int64

	alertStates common.Stats
	alerts      *common.AlertBucket

	mutex sync.RWMutex
}

func newNode(cluster *Cluster, origNode *as.Node) *Node {
	var host *as.Host
	if origNode != nil {
		host = origNode.GetHost()
	}

	lh := rrd.NewSimpleBucket(cluster.UpdateInterval(), 3600)

	return &Node{
		cluster:        cluster,
		origNode:       origNode,
		origHost:       host,
		status:         nodeStatus.On,
		namespaces:     map[string]*Namespace{},
		statsHistory:   map[string]*rrd.Bucket{},
		latencyHistory: lh,
		alertStates:    common.Stats{},
		alerts:         common.NewAlertBucket(db, 50),
	}
}

func (n *Node) valid() bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.origNode != nil && n.origNode.IsActive()
}

func (n *Node) update() error {
	defer n.notifyAboutChanges()

	if !n.valid() {
		n.setStatus(nodeStatus.Off)
		log.Warningf("Node %s is not active.", n.origHost)
		return nil
	}
	n.setStatus(nodeStatus.On)

	if err := n.updateNamespaceNames(); err != nil {
		return err
	}

	tm := time.Now()

	// retry 3 times
	info, err := n.RequestInfo(3, n.infoKeys()...)
	if err != nil {
		n.setStatus(nodeStatus.Off)
		return err
	}

	n.setInfo(common.Info(info))
	n.setConfig(n.InfoAttrs("get-config:").ToInfo("get-config:"))

	n.setStatus(nodeStatus.On)

	latencyMap, nodeLatency := n.parseLatencyInfo(info["latency:"])
	n.setNodeLatency(nodeLatency)

	nsAggStats := common.Stats{}
	nsAggCalcStats := common.Stats{}
	for _, ns := range n.namespaces {
		ns.update(n.InfoAttrs("namespace/"+ns.name, "sets/"+ns.name, "get-config:context=namespace;id="+ns.name))
		ns.updateIndexInfo(n.Indexes(ns.name))
		ns.updateLatencyInfo(latencyMap[ns.name])
		ns.aggStats(nsAggStats, nsAggCalcStats)

		// update node's server time
		n.setServerTime(ns.ServerTime().Unix())
	}

	stats := common.Info(info).ToInfo("statistics").ToStats()
	n.setStats(stats, nsAggStats, nsAggCalcStats)
	n.updateHistory()

	if !common.AMCIsProd() {
		log.Debugf("Updating Node: %v, objects: %v, took: %s", n.Id(), stats["objects"], time.Since(tm))
	}

	return nil
}

func (n *Node) notifyAboutChanges() {
	go n.updateNotifications()
}

func (n *Node) applyStatsToAggregate(stats, calcStats common.Stats) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	stats.AggregateStats(n.stats)
	calcStats.AggregateStats(n.nsAggCalcStats)
}

func (n *Node) applyNsStatsToAggregate(stats, calcStats map[string]common.Stats) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	for k, ns := range n.namespaces {
		if stats[k] == nil {
			stats[k] = common.Stats{}
		}
		if calcStats[k] == nil {
			calcStats[k] = common.Stats{}
		}

		ns.aggStats(stats[k], calcStats[k])
	}
}

func (n *Node) applyNsSetStatsToAggregate(stats map[string]map[string]common.Stats) map[string]map[string]common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	for nsName, ns := range n.namespaces {
		setsInfo := ns.SetsInfo()

		if stats[nsName] == nil {
			stats[nsName] = setsInfo
			continue
		}

		for setName, setInfo := range setsInfo {
			if stats[nsName][setName] == nil {
				stats[nsName][setName] = setInfo
				continue
			}

			stats[nsName][setName].AggregateStats(setInfo)
		}
	}

	return stats
}

func (n *Node) RequestInfo(reties int, cmd ...string) (result map[string]string, err error) {
	if len(cmd) == 0 {
		return map[string]string{}, nil
	}

	for i := 0; i < reties; i++ {
		result, err = n.origNode.RequestInfo(cmd...)
		if err == nil {
			return result, nil
		}
		// TODO: only retry for EOF or Timeout errors
	}

	return result, err
}

func (n *Node) updateNamespaceNames() error {
	namespaceListMap, err := n.RequestInfo(3, "namespaces")
	if err != nil {
		return err
	}

	namespacesStr := namespaceListMap["namespaces"]
	namespaces := strings.Split(namespacesStr, ";")

	n.mutex.Lock()
	defer n.mutex.Unlock()
	for _, ns := range namespaces {
		if n.namespaces[ns] == nil {
			n.namespaces[ns] = &Namespace{node: n, name: ns, statsHistory: map[string]*rrd.Bucket{}, latencyHistory: rrd.NewSimpleBucket(5, 3600)}
		}
	}

	return nil
}

func (n *Node) LatestLatency() map[string]common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.latestNodeLatency
}

func (n *Node) LatencySince(tms string) []map[string]common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	tm := n.ServerTime().Unix() - 30*60
	if len(tms) > 0 {
		if v, err := n.latencyDateTime(tms); err == nil {
			tm = v.Unix()
		}
	}

	vs := n.latencyHistory.ValuesSince(time.Unix(tm, 0))
	vsTyped := make([]map[string]common.Stats, len(vs))
	for i := range vs {
		if vIfc := vs[i]; vIfc != nil {
			if v, ok := vIfc.(*interface{}); ok {
				vsTyped[i] = (*v).(map[string]common.Stats)
			}
		}
	}

	return vsTyped
}

func (n *Node) LatestThroughput() map[string]map[string]*common.SinglePointValue {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	res := make(map[string]map[string]*common.SinglePointValue, len(n.statsHistory))
	for name, bucket := range n.statsHistory {
		val := bucket.LastValue()
		if val == nil {
			tm := time.Now().Unix()
			val = common.NewSinglePointValue(&tm, nil)
		}
		res[name] = map[string]*common.SinglePointValue{
			n.Address(): val,
		}
	}

	return res
}

func (n *Node) ThroughputSince(tm time.Time) map[string]map[string][]*common.SinglePointValue {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	res := make(map[string]map[string][]*common.SinglePointValue, len(n.statsHistory))
	for name, bucket := range n.statsHistory {
		vs := bucket.ValuesSince(tm)
		if len(vs) == 0 {
			tm := time.Now().Unix()
			vs = []*common.SinglePointValue{common.NewSinglePointValue(&tm, nil)}
		}
		res[name] = map[string][]*common.SinglePointValue{
			n.Address(): vs,
		}
	}

	return res
}

func (n *Node) updateHistory() {
	// this uses a RLock, so we put it before the locks to avoid deadlock
	latestLatencyReport, err := n.latestLatencyReportDateTime()
	tm := n.ServerTime().Unix()

	n.mutex.Lock()
	defer n.mutex.Unlock()

	recordedStats := []string{
		"stat_read_success", "stat_read_reqs",
		"stat_write_success", "stat_write_reqs",

		"batch_read_success", "batch_read_reqs",

		"scan_success", "scan_reqs",
		"query_success", "query_reqs",

		"xdr_read_success", "xdr_read_reqs",
		"xdr_write_success", "xdr_write_reqs",

		"udf_success", "udf_reqs",
	}

	for _, stat := range recordedStats {
		bucket := n.statsHistory[stat]
		if bucket == nil {
			bucket = rrd.NewBucket(n.cluster.UpdateInterval(), 3600, true)
			n.statsHistory[stat] = bucket
		}

		if n.nsAggCalcStats.Get(stat) != nil {
			bucket.Add(tm, n.nsAggCalcStats.TryFloat(stat, 0))
		} else if n.stats.Get(stat) != nil {
			bucket.Add(tm, n.stats.TryFloat(stat, 0))
		}
	}

	if err == nil && !latestLatencyReport.IsZero() {
		n.latencyHistory.Add(latestLatencyReport.Unix(), n.latestNodeLatency)
	}
}

func (n *Node) latestLatencyReportDateTime() (time.Time, error) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	var latTime string
	for _, stats := range n.latestNodeLatency {
		latTime = stats["timestamp"].(string)
		break
	}

	return n.latencyDateTime(latTime)
}

// adds date part to the latency time string
func (n *Node) latencyDateTime(ts string) (time.Time, error) {
	const layout = "2006-1-2 15:04:05"

	if len(ts) == 0 {
		// return zero value
		return time.Time{}, nil
	}

	// remove the timezone if exists
	if len(ts) > 8 {
		ts = ts[:8]
	}

	// means there is no server time yet
	st := n.ServerTime()
	if st.IsZero() {
		return st, nil
	}

	// beginning of day
	t := st.Truncate(24 * time.Hour)
	year, month, day := t.Date()

	t2, err := time.Parse(layout, fmt.Sprintf("%d-%d-%d ", year, month, day)+ts)
	if err != nil {
		log.Debug("DATE PARSE ERROR! ", err)
		return t2, err
	}

	// must not be in the future
	if t2.After(st) {
		// format for yesterday
		time.Parse(layout, fmt.Sprintf("%d-%d-%d ", year, month, day-1)+ts)
	}
	return t2, nil
}

func (n *Node) XdrEnabled() bool {
	return n.StatsAttr("xdr_uptime") != nil
}

func (n *Node) XdrConfig() common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.latestInfo.ToInfo("get-config:context=xdr").ToStats()
}

func (n *Node) XdrStats() common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.latestInfo.ToInfo("statistics/xdr").ToStats()
}

func (n *Node) XdrStatus() XDRStatus {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	if n.latestConfig.TryString("enable-xdr", "") != "true" {
		return xdrStatus.Off
	}
	return xdrStatus.On
}

func (n *Node) SwitchXDR(on bool) error {
	return n.SetXDRConfig("enable-xdr", on)
}

func (n *Node) SetXDRConfig(name string, value interface{}) error {
	cmd := fmt.Sprintf("set-config:context=xdr;%s=%v", name, value)

	res, err := n.origNode.RequestInfo(cmd)
	if err != nil {
		return err
	}

	errMsg, exists := res[cmd]
	if exists && strings.ToLower(errMsg) != "ok" {
		return errors.New(errMsg)
	}

	return nil
}

func (n *Node) infoKeys() []string {
	res := []string{"node", "statistics", "features",
		"cluster-generation", "partition-generation", "build_time",
		"edition", "version", "build", "build_os", "bins", "jobs:",
		"sindex", "udf-list", "latency:", "get-config:", "cluster-name",
	}

	if n.Enterprise() {
		res = append(res, "get-dc-config", "get-config:context=xdr", "statistics/xdr")
	}

	// add namespace stat requests
	for ns, _ := range n.namespaces {
		res = append(res, "namespace/"+ns)
		res = append(res, "sets/"+ns)
		res = append(res, "get-config:context=namespace;id="+ns)
	}

	return res
}

func (n *Node) NamespaceByName(ns string) *Namespace {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.namespaces[ns]
}

func (n *Node) setStats(stats, nsStats, nsCalcStats common.Stats) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	// alias stats
	stats["queue"] = stats.TryInt("tsvc_queue", 0)
	stats["cluster_name"] = n.latestInfo.TryString("cluster-name", "")
	stats["xdr_read_reqs"] =
		stats.TryInt("xdr_read_success", 0) +
			stats.TryInt("xdr_read_error", 0) +
			stats.TryInt("xdr_read_notfound", 0)
	if v := stats.Get("xdr_read_success"); v != nil {
		stats["free_dlog_pct"] = stats.Get("dlog_free_pct")
		stats["stat_recs_logged"] = stats.Get("dlog_logged")
		stats["stat_dlog_recs_overwritten"] = stats.Get("dlog_overwritten_error")
		stats["stat_recs_linkdown_processed"] = stats.Get("dlog_processed_link_down")
		stats["stat_recs_localprocessed"] = stats.Get("dlog_processed_main")
		stats["stat_recs_replprocessed"] = stats.Get("dlog_processed_replica")
		stats["stat_recs_relogged"] = stats.Get("dlog_relogged")
		stats["used_recs_dlog"] = stats.Get("dlog_used_objects")
		stats["local_recs_notfound"] = stats.Get("xdr_read_not_found")
		stats["failednode_sessions_pending"] = stats.Get("xdr_active_failed_node_sessions")
		stats["linkdown_sessions_pending"] = stats.Get("xdr_active_link_down_sessions")
		stats["esmt_ship_avg_comp_pct"] = stats.Get("xdr_ship_compression_avg_pct")
		stats["hotkeys_fetched"] = stats.Get("xdr_hotkey_fetch")
		stats["noship_recs_hotkey"] = stats.Get("xdr_hotkey_skip")
		stats["stat_recs_inflight"] = stats.Get("xdr_ship_inflight_objects")
		stats["stat_recs_outstanding"] = stats.Get("xdr_ship_outstanding_objects")
		stats["err_recs_dropped"] = stats.Get("xdr_queue_overflow_error")
		stats["read_threads_avg_processing_time_pct"] = stats.Get("xdr_read_active_avg_pct")
		stats["local_recs_error"] = stats.Get("xdr_read_error")
		stats["read_threads_avg_waiting_time_pct"] = stats.Get("xdr_read_idle_avg_pct")
		stats["local_recs_fetch_avg_latency"] = stats.Get("xdr_read_latency_avg")
		stats["dispatch_request_queue_used_pct"] = stats.Get("xdr_read_reqq_used_pct")
		stats["dispatch_request_queue_used"] = stats.Get("xdr_read_reqq_used")
		stats["dispatch_response_queue_used"] = stats.Get("xdr_read_respq_used")
		stats["local_recs_fetched"] = stats.Get("xdr_read_success")
		stats["transaction_queue_used_pct"] = stats.Get("xdr_read_txnq_used_pct")
		stats["transaction_queue_used"] = stats.Get("xdr_read_txnq_used")
		stats["stat_recs_relogged_incoming"] = stats.Get("xdr_relogged_incoming")
		stats["stat_recs_relogged_outgoing"] = stats.Get("xdr_relogged_outgoing")
		stats["esmt_bytes_shipped"] = stats.Get("xdr_ship_bytes")
		stats["xdr_deletes_shipped"] = stats.Get("xdr_ship_delete_success")
		stats["err_ship_server"] = stats.Get("xdr_ship_destination_error")
		stats["latency_avg_ship"] = stats.Get("xdr_ship_latency_avg")
		stats["err_ship_client"] = stats.Get("xdr_ship_source_error")
		stats["stat_recs_shipped_ok"] = stats.Get("xdr_ship_success")
		stats["stat_recs_shipped"] = stats.Get("xdr_ship_success")
		stats["cur_throughput"] = stats.Get("xdr_throughput")
		stats["noship_recs_uninitialized_destination"] = stats.Get("xdr_uninitialized_destination_error")
		stats["noship_recs_unknown_namespace"] = stats.Get("xdr_unknown_namespace_error")

		stats["esmt-bytes-shipped"] = stats.Get("esmt_bytes_shipped")
		stats["xdr-uptime"] = stats.Get("xdr_uptime")
		stats["free-dlog-pct"] = stats.Get("free_dlog_pct")
		// stats["xdr_timelag"] = stats.Get("timediff_lastship_cur_secs")
	}

	n.stats = stats
	n.nsAggStats = nsStats
	n.nsAggCalcStats = nsCalcStats
}

func (n *Node) setConfig(stats common.Info) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.latestConfig = stats.ToStats()
}

func (n *Node) setNodeLatency(stats map[string]common.Stats) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.latestNodeLatency = stats
}

func (n *Node) ConfigAttrs(names ...string) common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	var res common.Stats
	if len(names) == 0 {
		res = make(common.Stats, len(n.latestConfig))
		for name, value := range n.latestConfig {
			res[name] = value
		}
	} else {
		res = make(common.Stats, len(names))
		for _, name := range names {
			res[name] = n.latestConfig[name]
		}
	}
	return res
}

func (n *Node) ConfigAttr(name string) interface{} {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.latestConfig[name]
}

func (n *Node) SetServerConfig(context string, config map[string]string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	cmd := "set-config:context=" + context
	for parameter, value := range config {
		cmd += fmt.Sprintf(";%s=%s", parameter, value)
	}

	res, err := n.RequestInfo(3, cmd)
	if err != nil {
		return err
	}

	errMsg, exists := res[cmd]
	if exists && strings.ToLower(errMsg) == "ok" {
		return nil
	}

	return errors.New(errMsg)
}

func (n *Node) setInfo(stats common.Info) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.oldInfo = n.latestInfo
	n.latestInfo = stats
}

func (n *Node) InfoAttrs(names ...string) common.Info {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	var res common.Info
	if len(names) == 0 {
		res = make(common.Info, len(n.latestInfo))
		for name, value := range n.latestInfo {
			res[name] = value
		}
	} else {
		res = make(common.Info, len(names))
		for _, name := range names {
			res[name] = n.latestInfo[name]
		}
	}
	return res
}

func (n *Node) InfoAttr(name string) string {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.latestInfo[name]
}

func (n *Node) AnyAttrs(names ...string) common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	var res common.Stats
	res = make(common.Stats, len(names))
	for _, name := range names {
		if v := n.stats.Get(name); v != nil {
			res[name] = v
		} else if v := n.nsAggStats.Get(name); v != nil {
			res[name] = v
		} else if v := n.nsAggCalcStats.Get(name); v != nil {
			res[name] = v
		} else if v := n.latestInfo.Get(name); v != nil {
			res[name] = v
		} else if v := n.latestConfig.Get(name); v != nil {
			res[name] = v
		}
	}
	return res
}

func (n *Node) StatsAttrs(names ...string) common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	var res common.Stats
	if len(names) == 0 {
		res = make(common.Stats, len(n.stats))
		for name, value := range n.stats {
			res[name] = value
		}
	} else {
		res = make(common.Stats, len(names))
		for _, name := range names {
			res[name] = n.stats.Get(name)
		}
	}
	return res
}

func (n *Node) StatsAttr(name string) interface{} {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.stats.Get(name)
}

func (n *Node) Status() NodeStatus {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.status
}

func (n *Node) Enterprise() bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return strings.Contains(strings.ToLower(n.latestInfo.TryString("edition", "")), "enterprise")
}

func (n *Node) VisibilityStatus() NodeVisibilityStatus {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	res := nodeVisibilityStatus.On
	if n.Status() == nodeStatus.On && (n.origNode == nil || !n.origNode.IsActive()) {
		res = nodeVisibilityStatus.Off
	}
	return res
}

func (n *Node) setStatus(status NodeStatus) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.status = status
}

func (n *Node) Build() string {
	return n.InfoAttr("build")
}

func (n *Node) Disk() common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return common.Stats{
		"used":             n.nsAggCalcStats["used-bytes-disk"],
		"free":             n.nsAggCalcStats["free-bytes-disk"],
		"used-bytes-disk":  n.nsAggCalcStats.TryInt("used-bytes-disk", 0),
		"free-bytes-disk":  n.nsAggCalcStats.TryInt("free-bytes-disk", 0),
		"total-bytes-disk": n.nsAggCalcStats.TryInt("total-bytes-disk", 0),
	}
}

func (n *Node) Memory() common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return common.Stats{
		"used":               n.nsAggCalcStats["used-bytes-memory"],
		"free":               n.nsAggCalcStats["free-bytes-memory"],
		"used-bytes-memory":  n.nsAggCalcStats.TryInt("used-bytes-memory", 0),
		"free-bytes-memory":  n.nsAggCalcStats.TryInt("free-bytes-memory", 0),
		"total-bytes-memory": n.nsAggCalcStats.TryInt("total-bytes-memory", 0),
	}
}

func (n *Node) DataCenters() map[string]common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	if _, exists := n.latestInfo["get-dc-config"]; !exists {
		return nil
	}

	dcs := n.latestInfo.ToStatsMap("get-dc-config", "DC_Name", ":")
	for k, stats := range dcs {
		nodes := strings.Split(stats.TryString("Nodes", ""), ",")
		for i := range nodes {
			nodes[i] = strings.Replace(nodes[i], "+", ":", -1)
		}
		stats["Nodes"] = common.DeleteEmpty(nodes)
		stats["namespaces"] = common.DeleteEmpty(strings.Split(stats["namespaces"].(string), ","))
		dcs[k] = stats

		if len(nodes) == 0 {
			delete(dcs, k)
		}
	}

	return dcs
}

func (n *Node) NamespaceList() []string {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	res := make([]string, 0, len(n.namespaces))
	for ns, _ := range n.namespaces {
		res = append(res, ns)
	}
	return common.StrUniq(res)
}

func (n *Node) Jobs() map[string]common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.latestInfo.ToStatsMap("jobs:", "trid", ":")
}

func (n *Node) Indexes(namespace string) map[string]common.Info {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	indexes := n.latestInfo.ToInfoMap("sindex", "indexname", ":")
	if namespace == "" {
		return indexes
	}

	result := map[string]common.Info{}
	for idxName, idxInfo := range indexes {
		if idxInfo["ns"] == namespace {
			result[idxName] = idxInfo
		}
	}

	return result
}

func (n *Node) NamespaceIndexes() map[string][]string {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	indexes := n.latestInfo.ToInfoMap("sindex", "indexname", ":")
	result := map[string][]string{}
	for idxName, idxInfo := range indexes {
		if idxInfo != nil && idxInfo["ns"] != "" {
			result[idxInfo["ns"]] = append(result[idxInfo["ns"]], idxName)
		}
	}

	for k, v := range result {
		result[k] = common.StrUniq(v)
	}

	return result
}

func (n *Node) UDFs() map[string]common.Stats {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.latestInfo.ToStatsMap("udf-list", "filename", ",")
}

func (n *Node) Address() string {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	h := n.origNode.GetHost()
	return h.Name + ":" + strconv.Itoa(h.Port)
}

func (n *Node) Host() string {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	h := n.origNode.GetHost()
	return h.Name
}

func (n *Node) Port() uint16 {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	h := n.origNode.GetHost()
	return uint16(h.Port)
}

func (n *Node) Id() string {
	return n.InfoAttr("node")
}

func (n *Node) ClusterName() string {
	return n.InfoAttr("cluster-name")
}

func (n *Node) Bins() common.Stats {
	return parseBinInfo(n.InfoAttr("bins"))
}

func (n *Node) MigrationStats() common.Info {
	return n.InfoAttrs("migrate_msgs_sent", "migrate_msgs_recv", "migrate_progress_send",
		"migrate_progress_recv", "migrate_tx_objs", "migrate_rx_objs")
}

func parseBinInfo(s string) common.Stats {
	res := common.Stats{}
	for {
		i := strings.Index(s, ":bin_names=")
		ns := s[:i]
		s = s[i+11:]

		i = strings.Index(s, ",bin_names_quota=")
		bnCount, _ := strconv.Atoi(s[:i])
		s = s[i+17:]

		noBins := false
		i = strings.Index(s, ",")
		if i < 0 {
			i = strings.Index(s, ";")
			noBins = true
		}
		bnQuota, _ := strconv.Atoi(s[:i])
		s = s[i+1:]

		binNames := []string{}
		if !noBins {
			n := 0
			for {
				i = strings.Index(s[n:], ";")
				sep := strings.Index(s[n:], ",")
				// `;` is part of the bin name
				n += i
				if i < sep {
					continue
				}
				break
			}

			binNames = strings.Split(s[:n], ",")
		}

		res[ns] = common.Stats{
			"bin_names":       bnCount,
			"bin_names_quota": bnQuota,
			"bins":            binNames,
		}

		if len(s) > i {
			s = s[i+1:]
		} else {
			break
		}
	}

	return res
}

func (n *Node) setServerTime(tm int64) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if tm > n.serverTime {
		n.serverTime = tm
	}
}

func (n *Node) ServerTime() time.Time {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return time.Unix(n.serverTime, 0)
}

func (n *Node) parseLatencyInfo(s string) (map[string]common.Stats, map[string]common.Stats) {
	ip := common.NewInfoParser(s)

	//typical format is {test}-read:10:17:37-GMT,ops/sec,>1ms,>8ms,>64ms;10:17:47,29648.2,3.44,0.08,0.00;

	nodeStats := map[string]common.Stats{}
	res := map[string]common.Stats{}
	for {
		if err := ip.Expect("{"); err != nil {
			// it's an error string, read to next section
			if _, err := ip.ReadUntil(';'); err != nil {
				break
			}
			continue
		}

		ns, err := ip.ReadUntil('}')
		if err != nil {
			break
		}

		if err := ip.Expect("-"); err != nil {
			break
		}

		op, err := ip.ReadUntil(':')
		if err != nil {
			break
		}

		timestamp, err := ip.ReadUntil(',')
		if err != nil {
			break
		}

		if _, err := ip.ReadUntil(','); err != nil {
			break
		}

		bucketsStr, err := ip.ReadUntil(';')
		if err != nil {
			break
		}
		buckets := strings.Split(bucketsStr, ",")

		_, err = ip.ReadUntil(',')
		if err != nil {
			break
		}

		opsCount, err := ip.ReadFloat(',')
		if err != nil {
			break
		}

		valBucketsStr, err := ip.ReadUntil(';')
		if err != nil {
			break
		}
		valBuckets := strings.Split(valBucketsStr, ",")
		valBucketsFloat := make([]float64, len(valBuckets))
		for i := range valBuckets {
			valBucketsFloat[i], _ = strconv.ParseFloat(valBuckets[i], 64)
		}

		// calc precise in-between percents
		lineAggPct := float64(0)
		for i := len(valBucketsFloat) - 1; i > 0; i-- {
			lineAggPct += valBucketsFloat[i]
			valBucketsFloat[i-1] = math.Max(0, valBucketsFloat[i-1]-lineAggPct)
		}

		if len(buckets) != len(valBuckets) {
			log.Errorf("Error parsing latency values for node: `%s`. Bucket mismatch: buckets: `%s`, values: `%s`.", n.Address(), bucketsStr, valBucketsStr)
			break
		}

		stats := common.Stats{
			"tps":        opsCount,
			"timestamp":  timestamp,
			"buckets":    buckets,
			"valBuckets": valBucketsFloat,
		}

		if res[ns] == nil {
			res[ns] = common.Stats{
				op: stats,
			}
		} else {
			res[ns][op] = stats
		}

		// calc totals
		if nstats := nodeStats[op]; nstats == nil {
			nodeStats[op] = stats
		} else {
			// nstats := nstatsIfc.(common.Stats)
			if timestamp > nstats["timestamp"].(string) {
				nstats["timestamp"] = timestamp
			}
			nstats["tps"] = nstats["tps"].(float64) + opsCount
			nBuckets := nstats["buckets"].([]string)
			if len(buckets) > len(nBuckets) {
				nstats["buckets"] = append(nBuckets, buckets[len(nBuckets):]...)
				nstats["valBuckets"] = append(nstats["valBuckets"].([]float64), make([]float64, len(nBuckets))...)
			}

			nValBuckets := nstats["valBuckets"].([]float64)
			for i := range buckets {
				nValBuckets[i] += valBucketsFloat[i] * opsCount
			}
			nstats["valBuckets"] = nValBuckets

			nodeStats[op] = nstats
		}
	}

	return res, nodeStats
}

func (n *Node) updateNotifications() error {
	n.CheckStatus()
	n.CheckClusterVisibility()
	n.CheckTransactionQueue()
	n.CheckFileDescriptors()
	n.CheckDiskSpace()
	n.CheckMemory()

	return nil
}