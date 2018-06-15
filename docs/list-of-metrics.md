# mysql-monitoring-release Metrics

The metrics name will be pre-fixed by the value configured in the `mysql-metrics.origin` property on the `mysql-metrics` bosh job.

## MySQL Metrics

Metrics from MySQL, for use with both [cf-mysql-release](https://github.com/cloudfoundry/cf-mysql-release) and [pxc-release](https://github.com/cloudfoundry-incubator/pxc-release) deployed in all topologies

|Metric Name | Description | Units |
|------------|--------------------------------|-------------------------- |
| `available` | Indicates if the local database server is available and responding. | boolean |
| `innodb/buffer_pool_pages_free` | The number of free pages in the InnoDB Buffer Pool. | pages |
| `innodb/buffer_pool_pages_total` | The total number of pages in the InnoDB Buffer Pool. | pages |
| `innodb/buffer_pool_pages_data` | **TODO** | pages |
| `innodb/buffer_pool_used` | The number of used pages in the InnoDB Buffer Pool. | pages |
| `innodb/buffer_pool_utilization` | The utilization of the InnoDB Buffer Pool. | fraction |
| `innodb/current_row_locks` | The number of current row locks. | locks |
| `innodb/data_read` | The rate of data read. | reads/second |
| `innodb/data_written` | The rate of data written. | writes/second |
| `innodb/mutex_os_waits` | The rate of mutex OS waits. | events/second |
| `innodb/mutex_spin_rounds` | The rate of mutex spin rounds. | events/second |
| `innodb/mutex_spin_waits` | The rate of mutex spin waits. | events/second |
| `innodb/os_log_fsyncs` | The rate of fsync writes to the log file. | writes/second |
| `innodb/row_lock_time` | Time spent in acquiring row locks. | milliseconds |
| `innodb/row_lock_waits` | The number of times per second a row lock had to be waited for. | events/second |
| `innodb/row_lock_current_waits` | **TODO** | **TODO** what does "lock" mean? |
| `net/connections` | The rate of connections to the server. | connection/second |
| `net/max_used_connections` | The maximum number of connections that have been in use simultaneously since the server started. | connections |
| `performance/com_delete` | The rate of delete statements. | queries/second |
| `performance/com_delete_multi` | The rate of delete-multi statements. | queries/second |
| `performance/com_insert` | The rate of insert statements. | query/second |
| `performance/com_insert_select` | The rate of insert-select statements. | queries/second |
| `performance/com_replace_select` | The rate of replace-select statements. | queries/second |
| `performance/com_select` | The rate of select statements. | queries/second |
| `performance/com_update` | The rate of update statements. | queries/second |
| `performance/com_update_multi` | The rate of update-multi. | queries/second |
| `performance/cpu_time` | **TODO** | seconds |
| `performance/created_tmp_disk_tables` | The rate of internal on-disk temporary tables created by second by the server while executing statements. | table/second |
| `performance/created_tmp_files` | The rate of temporary files created by second. | files/second |
| `performance/created_tmp_tables` | The rate of internal temporary tables created by second by the server while executing statements. | tables/second |
| `performance/kernel_time` | Percentage of CPU time spent in kernel space by MySQL. | percent |
| `performance/key_cache_utilization` | The key cache utilization ratio. | fraction |
| `performance/open_files` | The number of open files. | files |
| `performance/open_tables` | The number of tables that are open. | tables |
| `performance/qcache_hits` | The rate of query cache hits. | hits/second |
| `performance/questions` | The rate of statements executed by the server. | queries/second |
| `performance/queries` | The rate of statements executed by the server, excluding `COM_PING` and `COM_STATISTICS`. Differs from `Questions` in that it also counts statements executed within stored programs. | queries/second |
| `performance/slow_queries` | The rate of slow queries. | queries/second |
| `performance/table_locks_waited` | The total number of times that a request for a table lock could not be granted immediately and a wait was needed. | number |
| `performance/threads_connected` | The number of currently open connections. | connections |
| `performance/threads_running` | The number of threads that are not sleeping. | threads |
| `performance/max_connections` | The maximum permitted number of simultaneous client connections. | integer |
| `performance/open_files_limit` | The number of files that the operating system permits [ **mysqld** ](https://dev.mysql.com/doc/refman/5.6/en/mysqld.html "4.3.1 mysqld — The MySQL Server") to open. | integer |
| `performance/open_tables` | The number of tables that are open. | integer |
| `performance/open_table_definitions` | **TODO** | integer |
| `performance/opened_tables` | The number of tables that have been opened. | integer |
| `performance/opened_table_definitions` | The number of `.frm` files that have been cached. | integer |
| `performance/queries` | The number of statements executed by the service, which resets to zero when the MySQL process is restarted. | integer |
| `performance/queries_delta` | The change in the `/performance/queries` metric since the last time it was emitted. | integer greater than zero |
| `variables/max_connections` | **TODO** | connections |
| `variables/open_files_limit` | **TODO** | files |
| `variables/read_only` | **TODO** | boolean |

## System Metrics
Metrics calculated from system calls, for use with both [cf-mysql-release](https://github.com/cloudfoundry/cf-mysql-release) and [pxc-release](https://github.com/cloudfoundry-incubator/pxc-release) deployed in all topologoies

Metric Name | Description | Units |
|------------|--------------------------------|-------------------------- |
| `performance/cpu_utilization_percent` | The percent of the CPU in the use by all processes on the MySQL node. | percent utilization, from 0-100 |
| `system/ephemeral_disk_free` | The number of KB available on the ephemeral disk. | KB |
| `system/ephemeral_disk_inodes_free` | The number of inodes available on the ephemeral disk. | count |
| `system/ephemeral_disk_inodes_used` | The number of inodes used on the ephemeral disk. | count |
| `system/ephemeral_disk_inodes_used_percent` | The percentage of ephemeral disk inodes used by both the system and user applications. | percent |
| `system/ephemeral_disk_used` | The number of KB used on the ephemeral disk. | KB |
| `system/ephemeral_disk_used_percent` | The percentage of ephemeral disk used by both the system and user applications. | percent |
| `system/persistent_disk_free` | The number of KB available on the persistent disk. | KB |
| `system/persistent_disk_inodes_free` | The number of inodes available on the persistent disk. | count |
| `system/persistent_disk_inodes_used` | The number of inodes used on the persistent disk. | count |
| `system/persistent_disk_inodes_used_percent` | The percentage of persistent disk inodes used by both the system and user applications. | percent |
| `system/persistent_disk_used` | The number of KB used on the persistent disk. | KB |
| `system/persistent_disk_used_percent` | The percentage of persistent disk used by both the system and user applications. | percent |

## Galera Metrics
Useful when deploying [cf-mysql-release](https://github.com/cloudfoundry/cf-mysql-release) or [pxc-release](https://github.com/cloudfoundry-incubator/pxc-release) in a galera topology

Metric Name | Description | Units |
|------------|--------------------------------|-------------------------- |
| `galera/wsrep_ready` | Shows whether the node can accept queries. | boolean |
| `galera/wsrep_cluster_size` | The current number of nodes in the Galera cluster. | node |
| `galera/wsrep_cluster_status` | Shows the primary status of the cluster component that the node is in. | State ID.<br /> Values are Primary = 1, Non-primary = 0, Disconnected = -1 (See: [https://mariadb.com/kb/en/mariadb/galera-cluster-status-variables/)](https://mariadb.com/kb/en/mariadb/galera-cluster-status-variables/ ) |
| `galera/wsrep_flow_control_paused` | The fraction of time since the last mysql start or FLUSH STATUS command that replication was paused due to flow control. This is a measure of how much replication lag is slowing down the cluster. | float |
| `galera/wsrep_flow_control_sent` | Number of FC_PAUSE (flow control pause) events sent by this node. Unlike most status variables, the counter for this one does not reset every time you run the query. | int |
| `galera/wsrep_flow_control_received` | Number of FC_PAUSE (flow control pause) events received by this node. This includes FC_PAUSE events sent by this node (it receives from itself). Unlike most status variables, the counter for this one does not reset every time you run the query. | int |
| `galera/wsrep_local_recv_queue_avg` | Shows the average size of the local received queue since the last status query. | float |
| `galera/wsrep_local_send_queue_avg` | Shows the average size of the local sent queue since the last status query. | float |
| `galera/wsrep_local_index` | This node index in the cluster (base 0). | int |
| `galera/wsrep_local_state` | This is the node's local state | float |


## Leader Follower Metrics
Useful when deploying [pxc-release](https://github.com/cloudfoundry-incubator/pxc-release) in a leader-follower topology

Metric Name | Description | Units |
|------------|--------------------------------|-------------------------- |
| `follower/is_follower` | **TODO** | boolean|
| `follower/relay_log_space` | **TODO** | bytes|
| `follower/seconds_behind_master` | **TODO** | seconds |
| `follower/seconds_since_leader_heartbeat` | **TODO** | seconds|
| `follower/slave_io_running` | **TODO** | boolean |
| `follower/slave_sql_running` | **TODO** | boolean |
| `rpl_semi_sync_master_no_tx` | **TODO** | **TODO** what does the integer mean? |
| `rpl_semi_sync_master_tx_avg_wait_time` | **TODO** | microsecond |
| `rpl_semi_sync_master_wait_sessions` | **TODO** | **TODO** what does the integer mean? |

## Broker Metrics
Can be implemented when deploying mysql releases with a service broker

Metric Name | Description | Units |
|------------|--------------------------------|-------------------------- |
| `broker/disk_allocated_service_plans` | The number of MB allocated by the broker for all service plans, current and allocated. | MB |
