# mysql-monitoring-release Metrics

The metrics name will be pre-fixed by the value configured in the `mysql-metrics.origin` property on the `mysql-metrics` bosh job.

## MySQL Metrics

Metrics from MySQL, for use with both [cf-mysql-release](https://github.com/cloudfoundry/cf-mysql-release) and [pxc-release](https://github.com/cloudfoundry-incubator/pxc-release) deployed in all topologies. Many of these metrics are total counts (usually since the server started or the last `flush status`), and you may find it useful to compute averages or deltas on the client side.

|Emitted Metric Name | Mysql Variable or Status Name| Description | Units |
|------------|-----| ---------------------------|-------------------------- |
| `available` | | Indicates if the local database server is available and responding. | boolean |
| `innodb/buffer_pool_pages_free` | [innodb_buffer_pool_pages_free](https://dev.mysql.com/doc/refman/5.6/en/server-status-variables.html#statvar_Innodb_buffer_pool_pages_free) | The number of free pages in the InnoDB Buffer Pool. | pages |
| `innodb/buffer_pool_pages_total` | [innodb_buffer_pool_pages_total](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_buffer_pool_pages_total)| The total number of pages in the InnoDB Buffer Pool. | pages |
| `innodb/buffer_pool_pages_data` | [innodb/buffer_pool_pages_data](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_buffer_pool_pages_data) | | pages |
| `innodb/data_read` | [innodb_data_read](https://dev.mysql.com/doc/refman/8.0/en/server-status-variables.html#statvar_Innodb_data_read) | The rate of data read. | reads/second |
| `innodb/data_written` | [innodb_data_written](https://dev.mysql.com/doc/refman/8.0/en/server-status-variables.html#statvar_Innodb_data_written) | The rate of data written. | writes/second |
| `innodb/mutex_os_waits` | [innodb_mutex_os_waits](https://mariadb.com/kb/en/library/xtradbinnodb-server-status-variables/#innodb_mutex_os_waits) | The rate of mutex OS waits. Emitted only by cf-mysql-release.| events/second |
| `innodb/mutex_spin_rounds` | [innodb_mutex_spin_rounds](https://mariadb.com/kb/en/library/xtradbinnodb-server-status-variables/#innodb_mutex_spin_rounds) | The rate of mutex spin rounds. Emitted only by cf-mysql-release. | events/second |
| `innodb/mutex_spin_waits` | [innodb_mutex_spin_waits](https://mariadb.com/kb/en/library/xtradbinnodb-server-status-variables/#innodb_mutex_spin_waits)| The rate of mutex spin waits. Emitted only by cf-mysql-release. | events/second |
| `innodb/os_log_fsyncs` | [innodb_os_log_fsyncs](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_os_log_fsyncs) | The rate of fsync writes to the log file. | writes/second |
| `innodb/row_lock_time` | [innodb_row_lock_time](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_row_lock_time) | Time spent in acquiring row locks. | milliseconds |
| `innodb/row_lock_waits` | [innodb_row_lock_waits](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_row_lock_waits) | The number of times per second a row lock had to be waited for. | events/second |
| `innodb/row_lock_current_waits` | [innodb_row_lock_current_waits](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_row_lock_current_waits) | | locks |
| `net/connections` | [connections](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Connections) | | connections|
| `net/max_used_connections` | [max_used_connections](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Max_used_connections) | The maximum number of connections that have been in use simultaneously since the server started. | connections |
| `performance/com_delete` | [com_delete](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of delete statements. | queries |
| `performance/com_delete_multi` | [com_delete_multi](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of delete-multi statements. | queries |
| `performance/com_insert` | [com_insert](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of insert statements. | query |
| `performance/com_insert_select` | [com_insert_select](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of insert-select statements. | queries |
| `performance/com_replace_select` | [com_replace_select](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of replace-select statements. | queries |
| `performance/com_select` | [com_select](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of select statements. | queries |
| `performance/com_update` | [com_update](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of update statements. | queries |
| `performance/com_update_multi` | [com_update_multi](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of update-multi. | queries |
| `performance/cpu_time` | [cpu_time](https://mariadb.com/kb/en/library/server-status-variables/#cpu_time) | Emitted only by cf-mysql-release. | |
| `performance/created_tmp_disk_tables` | [created_tmp_disk_tables](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Created_tmp_disk_tables) | The rate of internal on-disk temporary tables created by the server while executing statements. | table |
| `performance/created_tmp_files` | [created_tmp_files](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Created_tmp_files) | The rate of temporary files created. | files|
| `performance/created_tmp_tables` | [created_tmp_tables](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Created_tmp_tables) | The rate of internal temporary tables created by the server while executing statements. | tables |
| `performance/open_files` | [open_files](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Open_files) | The number of open files. | files |
| `performance/open_tables` | [open_tables](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Open_tables) | The number of tables that are currently open. | integer |
| `performance/open_table_definitions` | [open_table_definitions](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Open_table_definitions) | The number of currently cached table definitions| integer |
| `performance/opened_tables` | [opened_tables](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Opened_tables) | The number of tables that have been opened. | integer |
| `performance/opened_table_definitions` | [opened_table_definitions](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Opened_table_definitions) | The number of `.frm` files that have been cache. | integer |
| `performance/qcache_hits` | [qcache_hits](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Qcache_hits) | The number of query cache hits. | hits |
| `performance/questions` | [questions](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Questions) | The number of statements executed by the server. | queries |
| `performance/queries` | [queries](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Queries) | The number of statements executed by the server, excluding `COM_PING` and `COM_STATISTICS`. Differs from `Questions` in that it also counts statements executed within stored programs. | queries |
| `performance/queries_delta` | |  The change in the `/performance/queries` metric since the last time it was emitted. | integer greater than zero |
| `performance/slow_queries` | [slow_queries](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Slow_queries) | The number of slow queries. | queries |
| `performance/table_locks_waited` | [table_locks_waited](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Table_locks_waited) | The total number of times that a request for a table lock could not be granted immediately and a wait was needed. | number |
| `performance/threads_connected` | [threads_connected](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Threads_connected) | The number of currently open connections. | connections |
| `performance/threads_running` | [threads_running](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Threads_running) | The number of threads that are not sleeping. | threads |
| `variables/max_connections` | [max_connections](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_max_connections) | The maximum permitted number of simultaneous client connections. | connections |
| `variables/open_files_limit` | [open_files_limit](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_open_files_limit) |  The number of files that the operating system permits [ **mysqld** ](https://dev.mysql.com/doc/refman/5.6/en/mysqld.html "4.3.1 mysqld — The MySQL Server") to open. | files |
| `variables/read_only` | [read_only](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_read_only) | Whether the server is in read-only mode | boolean |

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

Metric Name | Galera Status Name | Description | Units |
|------------|---------|-----------------------|-------------------------- |
| `galera/wsrep_ready` | [wsrep_ready](http://galeracluster.com/documentation-webpages/galerastatusvariables.html#wsrep-ready) | Shows whether the node can accept queries. | boolean |
| `galera/wsrep_cluster_size` | [wsrep_cluster_size](http://galeracluster.com/documentation-webpages/galerastatusvariables.html#wsrep-cluster-size) | The current number of nodes in the Galera cluster. | node |
| `galera/wsrep_cluster_status` | [wsrep_cluster_status](http://galeracluster.com/documentation-webpages/galerastatusvariables.html#wsrep-cluster-status) | Shows the primary status of the cluster component that the node is in. | State ID.<br /> Values are Primary = 1, Non-primary = 0, Disconnected = -1 (See: [https://mariadb.com/kb/en/mariadb/galera-cluster-status-variables/)](https://mariadb.com/kb/en/mariadb/galera-cluster-status-variables/ ) |
| `galera/wsrep_flow_control_paused` | [wsrep_flow_control_paused](http://galeracluster.com/documentation-webpages/galerastatusvariables.html#wsrep-flow-control-paused) | The fraction of time since the last mysql start or FLUSH STATUS command that replication was paused due to flow control. This is a measure of how much replication lag is slowing down the cluster. | float |
| `galera/wsrep_flow_control_sent` | [wsrep_flow_control_sent](http://galeracluster.com/documentation-webpages/galerastatusvariables.html#wsrep-flow-control-sent) | Number of FC_PAUSE (flow control pause) events sent by this node. Unlike most status variables, the counter for this one does not reset every time you run the query. | int |
| `galera/wsrep_flow_control_recv` | [wsrep_flow_control_recv](http://galeracluster.com/documentation-webpages/galerastatusvariables.html#wsrep-flow-control-recv) | Number of FC_PAUSE (flow control pause) events received by this node. This includes FC_PAUSE events sent by this node (it receives from itself). Unlike most status variables, the counter for this one does not reset every time you run the query. | int |
| `galera/wsrep_local_recv_queue_avg` | [wsrep_local_recv_queue_avg](http://galeracluster.com/documentation-webpages/galerastatusvariables.html#wsrep-local-recv-queue-avg) | Shows the average size of the local received queue since the last status query. | float |
| `galera/wsrep_local_send_queue_avg` | [wsrep_local_send_queue_avg](http://galeracluster.com/documentation-webpages/galerastatusvariables.html#wsrep-local-send-queue-avg) | Shows the average size of the local sent queue since the last status query. | float |
| `galera/wsrep_local_index` | [wsrep_local_index](http://galeracluster.com/documentation-webpages/galerastatusvariables.html#wsrep-local-index) | This node index in the cluster (base 0). | int |
| `galera/wsrep_local_state` | [wsrep_local_state](http://galeracluster.com/documentation-webpages/galerastatusvariables.html#wsrep-local-state) | This is the node's local state | float |


## Leader Follower Metrics
Useful when deploying [pxc-release](https://github.com/cloudfoundry-incubator/pxc-release) in a leader-follower topology

Metric Name | Mysql Status or Variable Name| Description | Units |
|------------|---------|-------------------------------|-------------------------- |
| `follower/is_follower` | |  True if the server is following another | boolean|
| `follower/relay_log_space` | [relay_log_space](https://dev.mysql.com/doc/refman/5.7/en/show-slave-status.html) | | bytes|
| `follower/seconds_behind_master` | [seconds_behind_master](https://dev.mysql.com/doc/refman/5.7/en/show-slave-status.html) | | seconds |
| `follower/seconds_since_leader_heartbeat` | [seconds_since_leader_heartbeat](https://dev.mysql.com/doc/refman/5.7/en/show-slave-status.html) | | seconds|
| `follower/slave_io_running` | [slave_io_running](https://dev.mysql.com/doc/refman/5.7/en/show-slave-status.html) | | boolean |
| `follower/slave_sql_running` | [slave_sql_running](https://dev.mysql.com/doc/refman/5.7/en/show-slave-status.html) | | boolean |
| `rpl_semi_sync_master_no_tx` | [Rpl_semi_sync_master_no_tx](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Rpl_semi_sync_master_no_tx) | | commits |
| `rpl_semi_sync_master_tx_avg_wait_time` | [rpl_semi_sync_master_tx_avg_wait_time](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Rpl_semi_sync_master_tx_avg_wait_time) | | microsecond |
| `rpl_semi_sync_master_wait_sessions` | [rpl_semi_sync_master_wait_sessions](https://dev.mysql.com/doc/refman/5.7/en/show-slave-status.html) | | sessions |

## Broker Metrics
Can be implemented when deploying mysql releases with a service broker

Metric Name | Description | Units |
|------------|--------------------------------|-------------------------- |
| `broker/disk_allocated_service_plans` | The number of MB allocated by the broker for all service plans, current and allocated. | MB |
