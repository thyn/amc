/******************************************************************************
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
******************************************************************************/

define(["jquery"  ], function($){
     var VarDetails = {
        node : {
            //'objects' : ['Replicated Objs', false],
            'cluster_integrity' : ['If the node is experiencing any integrity issues like one-way network failures or disagreement on succession list etc ' , false],
            // 'err_out_of_space' : ['Number of writes resulting in disk out of space errors', false],
            // 'write_master' : ['Number of writes at the master', 'size'],
            // 'write_prole' : ['Number of writes at the prole', 'size'],
            'queue' : ['Number of pending requests waiting to be executed', 'number'],
            'system_swapping' : ['Amount of system swap memory that is being used', false],
            // 'err_write_fail_prole_unknown' : ['Number of prole write failures with uknown errors', 'number'],
            // 'stat_duplicate_operation' : ['Number of duplicate resolution operations ', false],
            'uptime' : ['Time in seconds since last server restart', 'time-seconds'],
            // 'nsup-threads' : ['Number of namespace-supervisor threads ', 'number'],
            // 'scan-priority' : ['Default scan-priority as per the configuration', 'number'],
            // 'storage-benchmarks' : ['If the storage benchmarks are enabled or not ', false],
            'system_free_mem_pct' : ['Percentage of free system memory', 'pct'],
            // 'replication-fire-and-forget' : ['replication-fire-and-forget', false],
            'free-pct-disk' : ['Percentage of disk free for the namespace', 'pct']
        },
        job : {
            //'objects' : ['Replicated Objs', false],
            'run-time' : ['' , false],
            'set' : ['' , false],
            'priority' : ['' , false],
            'net-io-bytes' : ['' , false],
            'recs-read' : ['' , false]

        },
        jobOldVersion : {
            //'objects' : ['Replicated Objs', false],
            'run_time' : ['' , false],
            'set' : ['' , false],
            'priority' : ['' , false],
            'net_io_bytes' : ['' , false],
            'recs_read' : ['' , false]

        },
        namespace : {
            'memory-size' : ['Max amount of memory for this namespace' , 'size'],
            'free-pct-memory' : ['Percentage of memory available', 'pct'],
            'available_pct' : ['This value determines the amount of space available for use by defrag' , 'pct' ],
            'max-void-time' : ['The time beyond which all the existing records in the namespace will expire', 'millisec'],
            'stop-writes' : ['Flag to indicate if stop-writes has triggered', false],
            'stop-writes-pct' : ['Disallow writes (except deletes) when either memory or disk is full more than specified percentage','pct'],
            'hwm-breached' : ['Flag to indicate if hwm is breached', false],
            'default-ttl' : ['Default time-to-live (in seconds) for a record from the time of creation or last updation. The record will be expired in the system beyond this time limit', 'sec'],
            'max-ttl' : ['Maximum TTL allowed in the server', 'sec'],
            'enable-xdr' : ['If the XDR is enabled for the namespace or not ', false],
            'single-bin' : ['Setting it true will disallow multiple columns for a record', false],
            'data-in-memory' : ['keep a copy of all data in memory always', false]
        },
        allStats : {//All stats for nodes, namespaces, xdr
            //Nodes
            'cluster_size':	'Size of the cluster',
            'objects':	'Total number of records in the cluster',
            'used-bytes-disk':	'Size of disk in use ( bytes )',
            'total-bytes-disk':	'Total size of disk ( bytes )',
            'free-pct-disk':	'Percentage of disk available',
            'used-bytes-memory':	'Size of memory in use (bytes)',
            'total-bytes-memory':	'Total Size of memory (bytes)',
            'free-pct-memory':	'Percentage of memory available',
            'stat_read_reqs':	'Number of total read requests',
            'stat_read_success':	'Number of read requests successful',
            'stat_read_errs_notfound':	'Number of read requests resulting in error : "key not found"',
            'stat_read_errs_other':	'Number of read requests resulting in other errors',
            'stat_read_latency_gt50':	'Number of times read requests took more than 50 ms',
            'stat_read_latency_gt100':	'Number of times read requests took more than 100 ms',
            'stat_read_latency_gt250':	'Number of times read requests took more than 250 ms',
            'stat_write_reqs':	'Number of total writes requested',
            'stat_write_success':	'Number of write requests successful',
            'stat_write_errs':	'Number of write requests resulting in errors',
            'stat_write_latency_gt50':	'Number of times write requests took more than 50 ms',
            'stat_write_latency_gt100':	'Number of times write requests took more than 100 ms',
            'stat_write_latency_gt250':	'Number of times write requests took more than 250 ms',
            'stat_rw_timeout':	'Number of read write requests failed because of timeout on the server side',
            'stat_proxy_reqs':	'Number of proxy requests',
            'stat_proxy_success':	'Number of proxy requests served successfully',
            'stat_proxy_errs':	'Number of proxy requests returning errors',
            'stat_proxy_latency_gt50':	'Number of times proxy requests took more than 50 ms',
            'stat_proxy_latency_gt100':	'Number of times proxy requests took more than 100 ms',
            'stat_proxy_latency_gt250':	'Number of times proxy requests took more than 250 ms',
            'stat_expired_objects':	'Number of objects expired',
            'stat_evicted_objects':	'Number of objects evicted',
            'stat_evicted_objects_time':	'Average expiry time(TTL) of the objects evicted in the last iteration',
            'stat_evicted_objects_min':	'Minimum TTL of the objects evicted in the last iteration',
            'stat_single_bin_records':	'Number of single bin records',
            'stat_zero_bin_records':	'Number of write requests that failed because of zero bin records',
            'stat_zero_bin_records_read':	'Number of read requests that failed because of zero bin records',
            'err_tsvc_requests':	'Number of failures where the execution is not even attempted',
            'err_out_of_space':	'Number of writes resulting in disk out of space errors',
            'err_duplicate_proxy_request':	'Number of duplicate proxy errors',
            'err_rw_request_not_found':	'Number of read/write transactions started but went amiss',
            'err_rw_pending_limit':	'Number of transactions failed on "hot keys" ',
            'err_rw_cant_put_unique':	'Number of duplicate transactions aborted',
            'err_write_empty_writes':	'Number of write failures where all the bin operations failed',
            'err_rcrb_reduce_gt5':	'Number of times index traversals took more than 5 ms (not used)',
            'err_rcrb_reduce_gt50':	'Number of times index traversals took more than 50 ms (not used)',
            'err_rcrb_reduce_gt100':	'Number of times index traversals took more than 100 ms (not used)',
            'err_rcrb_reduce_gt250':	'Number of times index traversals took more than 250 ms (not used)',
            'fabric_msgs_sent':	'Number of messages sent via fabric to other nodes',
            'fabric_msgs_rcvd':	'Number of messages received via fabric from other nodes',
            'migrate_msgs_sent':	'Number of migrate messages sent',
            'migrate_msgs_recv':	'Number of migrate messages received',
            'migrate_progress_send':	'Number of partitions pending to be migrated out of the node',
            'migrate_progress_recv':	'Number of partitions currently being migrated in to the node',
            'queue':	'Number of pending requests waiting to be executed',
            'transactions':	'Number of transactions running currently- includes all reads,writes,info commands etc.',
            'reaped_fds':	'Number of idle client fds closed',
            'scan_initiate':	'Number of scan requests initiated',
            'scan_pending':	'Number of scan requests pending',
            'proxy_initiate':	'Number of proxy requests initiated, as result of handling client request',
            'proxy_action':	'Number of proxy requests, received from other nodes',
            'proxy_retry':	'Number of proxy requests to nother nodes, retries',
            'proxy_retry_q_full':	'Number of proxy retries failed because fabric queue got full',
            'proxy_unproxy':	'Number of re-executions (from scratch) because of unavailability of proxy node',
            'proxy_retry_same_dest':	'Number of proxy retries to the same destination node',
            'proxy_retry_new_dest':	'Number of proxy retries to new nodes',
            'write_master':	'Number of writes at the master',
            'write_prole':	'Number of writes at the prole',
            'read_dup_master':	'Unused',
            'read_dup_prole':	'Number of requests sent for duplicate resolutions',
            'rw_err_dup_internal':	'Number of errors encountered during duplicate resolutions',
            'rw_err_dup_cluster_key':	'Number of errors encountered during duplicate resolutions because of cluster key mismatch',
            'rw_err_dup_send':	'Number of errors encountered during duplicate resolutions because of failure to send fabric messages',
            'rw_err_dup_write_internal':	'---> To be deleted <---',
            'rw_err_dup_write_cluster_key':	'---> To be deleted <---',
            'rw_err_write_internal':	'Number of write requests failed because of internal errors (code errors)',
            'rw_err_write_cluster_key':	'Unused',
            'rw_err_write_send':	'Number of prole write acknowledgements failed because of failure in sending fabric messages',
            'rw_err_ack_internal':	'Number of prole write acknowledgements failed due to internal errors',
            'rw_err_ack_nomatch':	'Number of prole write acknowledgements started but went amiss/ have mismatched information',
            'rw_err_ack_badnode':	'Number of acknowledgements from unexpected nodes',
            'waiting_transactions':	'Number of read/write (only. not scan,info etc) transactions queued currently',
            'tree_count':	'Number of index trees currently active in the node',
            'record_refs':	'Number of index records referenced currently',
            'record_locks':	'Number of index record locks currently active in the node',
            'migrate_tx_objs':	'Number of partitions pending to be migrated out of the node',
            'migrate_rx_objs':	'Number of partitions currently being migrated in to the node',
            'write_reqs':	'Number of entries currently in the write-transaction hash table',
            'storage_queue_full':	'Number of disk write failures due to storage space reservation issue',
            'storage_queue_delay':	'Number of times the disk write was delayed because there are few available write buffers',
            'partition_actual':	'Number of partitions for which this node is acting as master',
            'partition_replica':	'Number of partitions for which this node is acting as replica',
            'partition_desync':	'Number of partitions that are yet to be synced with the rest of the cluster nodes',
            'partition_absent':	'Number of partitions for which this node is not either master or replica',
            'partition_object_count':	'Total number of objects',
            'partition_ref_count':	'Number of partitions that are currently being read',
            'system_free_mem_pct':	'Percentage of free system memory',
            'system_swapping':	'Amount of system swap memory that is being used',
            'err_replica_null_node':	'Number of errors during cluster state exchange because of missing replica node information',
            'err_replica_non_null_node':	'Number of errors during cluster state exchange because of unexpected replica node information',
            'err_sync_copy_null_node':	'Number of errors during cluster state exchange because of missing general node information',
            'err_sync_copy_null_master':	'Number of errors during cluster state exchange because of missing master node information',
            'storage_defrag_records':	'Unused',
            'err_storage_defrag_fd_get':	'Unused',
            'storage_defrag_seek':	'Unused',
            'storage_defrag_read':	'Unused',
            'storage_defrag_bad_magic':	'Unused',
            'storage_defrag_sigfail':	'Unused',
            'storage_defrag_corrupt_record':	'Number of times the defrag thread encountered invalid records',
            'storage_defrag_wait':	'Number of times the defrag thread waited (slept)',
            'err_write_fail_prole_unknown':	'Number of prole write failures with uknown errors',
            'err_write_fail_prole_generation':	'Number of prole write failures because of generation mismatch',
            'err_write_fail_unknown':	'Number of write requests failed because of unknown errors',
            'err_write_fail_key_exists':	'Number of write requests that failed because the key already exists',
            'err_write_fail_generation':	'Number of write requests failed because of generation mismatch',
            'err_write_fail_bin_exists':	'Number of write requests resulting in error "bin exists" ',
            'err_write_fail_parameter':	'Number of writes request failing because of bad parameter from application code',
            'err_write_fail_noxds':	'Number of write requests failed because the XDR is not running',
            'uptime':	'Time in seconds since last server restart',
            'stat_write_errs_notfound':	'Number of errors returning key not found on a write request (mostly the case of delete)',
            'stat_write_errs_other':	'Number of other errors on a write request',
            'client_connections':	'Number of active client connections to the node.',
            'batch_errors':	'Number of errors for the batch requests',
            'batch_initiate':	'Number of batch requests initiated',
            'batch_queue':	'Number of batch queues',
            'batch_timeout':	'Number of batch requests timed out',
            'batch_tree_count':	'Count the number of trees (if you read this, congrats, i did not know what this means)',
            'err_write_fail_generation_xdr':	'Number of write requests from XDR that failed because of generation mismatch',
            'err_write_fail_prole_delete':	'Number of failures in deleting the replica because its not found',
            'heartbeat_received_foreign':	'Total number of heartbeats received from remote nodes. (If unchanging, could be a problem)',
            'heartbeat_received_self':	'Total number of heartbeats from this node received by this node (Will be 0 for mesh, but should be changing for multicast.)',
            'migrate_num_incoming_accepted':	'Total number of incoming migrates that have been accepted by this node.',
            'migrate_num_incoming_refused':	'Total number of incoming migrates this node has refused due to reaching the "migrate-max-num-incoming" limit',
            'stat_evicted_objects_max':	'--Deprecated--',
            'stat_evicted_set_objects':	'Number of evicted due to sets limits from config file.',
            'stat_proxy_reqs_xdr':	'Number of XDR operations that resulted in proxies',
            'stat_write_reqs_xdr':	'Number of write requests from XDR',
            'tscan_initiate':	'Number of scans initiated. New scan',
            'cluster_integrity':	'State to indicate the Cluster\'s paxos state',
            'err_write_fail_noxdr':	'Number of writes that were rejected because XDR was not running.',
            'stat_deleted_set_objects':	'Number of deleted set objects',
            'stat_nsup_deletes_not_shipped':	'Number of deletes not shipped if it is result of eviction/migrations etc',
            'tscan_pending':	'Number of scans pending',
            'info_queue':	'Size of the info queue',
            'stat_duplicate_operation':	' detect if there are duplicates, create a duplicate-request and start the duplicate-fetch phase',
            'stat_leaked_wblocks':	'Number of wblocks termed as leaked.',

            //Namespace
            'available_pct':	'This value determines the amount of space available for use by defrag. Should not be less than 20',
            'evict-pct':	'Maximum percentage of objects to be deleted during each iteration of eviction',
            'stop-writes-pct':	'Disallow writes (except deletes) when either memory or disk is full more than specified percentage',
            'high-water-memory':	'Evict data if memory utilization is greater than specified percentage',
            'high-water-disk':	'Evict data if disk utilization is greater than specified percentage',
            'low-water-pct':	'Expiration/Eviction thread will not do any activity if the used percentage of memory/disk is less than the specified limit',
            'allow_versions':	'Keep all the versions in case of conflict during migrations',
            'default-ttl':	'Default time-to-live (in seconds) for a record from the time of creation or last updation. The record will be expired in the system beyond this time limit',
            'conflict-resolution-policy':	'Current config resolution policy. (TTL or Generation)',
            'repl-factor':	'Number of copies of a record (including the master copy) maintained in the entire cluster',
            'current-time':	'Current time in epoch',
            'max-void-time':	'The maximum TTL in namespace.',
            'memory-size':	'Max amount of memory for this namespace',
            'expired-objects':	'Total number of records expired for the namespace in the cluster',
            'evicted-objects':	'Total number of records evicted for the namespace in the cluster',
            'stop-writes':	'Flag to indicate if stop-writes has triggered',
            'lwm-breached':	'Flag to indicate if lwm is breached',
            'hwm-breached':	'Flag to indicate if hwm is breached',
            'enable_xdr':	'If XDR is enabled for the namespace',
            'defrag-period':	'Defrag interval at which defrag will trigger',
            'defrag-max-blocks':	'Max blocks to defrag in each cycle',
            'defrag-lwm-pct':	'lwm for the blocks , if it goes below this value it will queue the block for defragging.',
            'defrag-startup-minimum':	'The minimum pct of available space required for the node to startup.',
            'dev':	'Devices configured for the namespace',
            'filesize':	'Filesize - default configured for 16GB -even when its SSD namespace',
            'single-bin':	'If the namespace is single-bin or multibin',
            'max-ttl':	'if any incoming write is greater than this value , it will reject writes',
            'write-smoothing-period':	'--Not-Supported--',
            'set-deleted-objects':	'Number of objects deleted by a set',
            'set-evicted-objects':	'Number of objects evicted by a set',
            'enable-xdr':	'Enable the namespace for XDR shipment',
            'high-water-disk-pct':	'High water mark percentage usage for disk',
            'high-water-memory-pct':	'High water mark percentage usage for memory',

            //XDR
            'stat_recs_logged':	'Number of records logged into digest log',
            'stat_recs_relogged':	'Number of relogged records.',
            'stat_recs_localprocessed':	'Number of records processed',
            'stat_recs_replprocessed':	'Number of records processed for the non-local node',
            'stat_recs_shipped':	'Number of records shipped to remote datacenter',
            'stat_recs_shipping':	'Number of records currently in shipping.',
            'stat_recs_outstanding':	'Number of overall outstanding records yet to be shipped',
            'stat_deletes_shipped':	'Number of deletes shipped',
            'stat_deletes_relogged':	'Number of deletes re-logged',
            'stat_deletes_canceled':	'Number of deletes canceled as we found records corresponding to the digest to be deleted',
            'local_recs_fetched':	'Number of records fetched from local cluster.',
            'noship_recs_genmismatch':	'Number of records not shipped due to generation mismatch. Remote data center does not have lower generation record',
            'noship_recs_unknown_namespace':	'Number of records not shipped due to XDR having no knowledge about the namespace to which the digest log belongs.',
            'noship_recs_notmaster':	'Number of records not shipped as they were from non-local node.',
            'err_ship_client':	'Number of connectivity errors while shipping data to remote data center',
            'err_ship_server':	'Number of write errors on the remote data center',
            'timediff_lastship_cur_secs': 'Time difference in seconds between last shipped record and current time. (Shipping Lag)',
            'xdr_timelag':	'Time difference in seconds between last shipped record and current time. (Shipping Lag)',
            'latency_avg_ship':	'Average shipping latency in milliseconds',
            'esmt-bytes-shipped':	'Estimated number of bytes shipped over the network to remote data center.',
            'used_recs_dlog':	'Number of record slots in the digest log in use',
            'total_recs_dlog':	'Total number of records slots available in digest log',
            'free_dlog_pct':	'Percentage of the digest log free and available for use',
            'xdr_uptime':	'Uptime in seconds.'


        }

    };

     return VarDetails;

});