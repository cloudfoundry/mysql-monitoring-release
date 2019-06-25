# mysql-monitoring-release Metrics

The metrics name will be pre-fixed by the value configured in the `mysql-metrics.source_id` property on the `mysql-metrics` bosh job.

<a name='mysql-metrics'>

## MySQL Metrics

Metrics from MySQL, for use with both [cf-mysql-release](https://github.com/cloudfoundry/cf-mysql-release) and [pxc-release](https://github.com/cloudfoundry-incubator/pxc-release) deployed in all topologies. Many of these metrics are total counts (usually since the server started or the last `flush status`), and you may find it useful to compute averages or deltas on the client side.

|Emitted Metric Name | Mysql Variable or Status Name| Description | Units |
|------------|-----| ---------------------------|-------------------------- |
| `available` | | Indicates if the local database server is available and responding. | boolean |
| `innodb/buffer_pool_pages_free` | [innodb_buffer_pool_pages_free](https://dev.mysql.com/doc/refman/5.6/en/server-status-variables.html#statvar_Innodb_buffer_pool_pages_free) | The amount of free space in the InnoDB Buffer Pool, in units of [pages](https://dev.mysql.com/doc/refman/5.7/en/glossary.html#glos_page). | pages |
| `innodb/buffer_pool_pages_total` | [innodb_buffer_pool_pages_total](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_buffer_pool_pages_total)| The total amount of free space in the InnoDB Buffer Pool, in units of [pages](https://dev.mysql.com/doc/refman/5.7/en/glossary.html#glos_page). | pages |
| `innodb/buffer_pool_pages_data` | [innodb/buffer_pool_pages_data](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_buffer_pool_pages_data) | The number of pages in the InnoDB buffer pool containing data. The number includes both dirty and clean pages.  | pages |
| `innodb/data_read` | [innodb_data_read](https://dev.mysql.com/doc/refman/8.0/en/server-status-variables.html#statvar_Innodb_data_read) | The amount of data read since the server started. | bytes |
| `innodb/data_written` | [innodb_data_written](https://dev.mysql.com/doc/refman/8.0/en/server-status-variables.html#statvar_Innodb_data_written) | The amount of data written the server started. | bytes |
| `innodb/mutex_os_waits` | [innodb_mutex_os_waits](https://mariadb.com/kb/en/library/xtradbinnodb-server-status-variables/#innodb_mutex_os_waits) | The number of mutex OS waits. Emitted only by cf-mysql-release.| count |
| `innodb/mutex_spin_rounds` | [innodb_mutex_spin_rounds](https://mariadb.com/kb/en/library/xtradbinnodb-server-status-variables/#innodb_mutex_spin_rounds) | The number of mutex spin rounds. Emitted only by cf-mysql-release. | count |
| `innodb/mutex_spin_waits` | [innodb_mutex_spin_waits](https://mariadb.com/kb/en/library/xtradbinnodb-server-status-variables/#innodb_mutex_spin_waits)| The number of mutex spin waits. Emitted only by cf-mysql-release. | count |
| `innodb/os_log_fsyncs` | [innodb_os_log_fsyncs](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_os_log_fsyncs) | The number of fsync() writes done to the InnoDB redo log files. | count|
| `innodb/row_lock_time` | [innodb_row_lock_time](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_row_lock_time) | Total time spent in acquiring row locks. | milliseconds |
| `innodb/row_lock_waits` | [innodb_row_lock_waits](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_row_lock_waits) | The number of times per second a row lock had to be waited for since the server started. | count |
| `innodb/row_lock_current_waits` | [innodb_row_lock_current_waits](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Innodb_row_lock_current_waits) | The number of row locks currently being waited for by operations on InnoDB tables. | count |
| `net/connections` | [connections](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Connections) | The number of connection attempts (successful or not) to the MySQL server. | connections|
| `net/max_used_connections` | [max_used_connections](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Max_used_connections) | The maximum number of connections that have been in use simultaneously since the server started. | connections |
| `performance/com_delete` | [com_delete](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of delete statements since the server started or the last `FLUSH STATUS`. | queries |
| `performance/com_delete_multi` | [com_delete_multi](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of delete-multi statements since the server started or the last `FLUSH STATUS`. Applies to DELETE statements that use multiple-table syntax. | queries |
| `performance/com_insert` | [com_insert](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of insert statements since the server started or the last `FLUSH STATUS`. | queries |
| `performance/com_insert_select` | [com_insert_select](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of insert-select statements since the server started or the last `FLUSH STATUS`. | queries |
| `performance/com_replace_select` | [com_replace_select](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of replace-select statements since the server started or the last `FLUSH STATUS`. | queries |
| `performance/com_select` | [com_select](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of select statements since the server started or the last `FLUSH STATUS`. If a query result is returned from query cache, the server increments the Qcache_hits status variable, not Com_select.| queries |
| `performance/com_update` | [com_update](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of update statements since the server started or the last `FLUSH STATUS`. | queries |
| `performance/com_update_multi` | [com_update_multi](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Com_xxx) | The number of update-multi statements since the server started or the last `FLUSH STATUS`. Applies to UPDATE statements that use multiple-table syntax.| queries |
| `performance/cpu_time` | [cpu_time](https://mariadb.com/kb/en/library/server-status-variables/#cpu_time) | Total CPU time used. Emitted only by cf-mysql-release. | |
| `performance/created_tmp_disk_tables` | [created_tmp_disk_tables](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Created_tmp_disk_tables) | The number of internal on-disk temporary tables created by the server while executing statements. | tables |
| `performance/created_tmp_files` | [created_tmp_files](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Created_tmp_files) | The number of temporary files created by mysqld. | files |
| `performance/created_tmp_tables` | [created_tmp_tables](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Created_tmp_tables) | The number of internal temporary tables created by the server while executing statements. | tables |
| `performance/open_files` | [open_files](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Open_files) | The number of regular files currently open, which were opened by the server. | files |
| `performance/open_tables` | [open_tables](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Open_tables) | The number of tables that are currently open. | tables |
| `performance/open_table_definitions` | [open_table_definitions](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Open_table_definitions) | The number of currently cached table definitions (`.frm` files). | count |
| `performance/opened_tables` | [opened_tables](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Opened_tables) | The number of tables that have been opened. | tables |
| `performance/opened_table_definitions` | [opened_table_definitions](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Opened_table_definitions) | The number of `.frm` files that have been cached. | integer |
| `performance/qcache_hits` | [qcache_hits](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Qcache_hits) | The number of query cache hits. The query cache and `qcache_hits` metric is deprecated as of MySQL 5.7.20. | hits |
| `performance/questions` | [questions](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Questions) | The number of statements executed by the server, since the server started or the last `FLUSH STATUS`. This includes only statements sent to the server by clients and not statements executed within stored programs, unlike the Queries variable. | count |
| `performance/queries` | [queries](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Queries) | The number of statements executed by the server, excluding `COM_PING` and `COM_STATISTICS`. Differs from `Questions` in that it also counts statements executed within stored programs. Not affected by `FLUSH STATUS`. | count |
| `performance/queries_delta` | |  The change in the `/performance/queries` metric since the last time it was emitted. | integer greater than or equal to zero |
| `performance/slow_queries` | [slow_queries](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Slow_queries) | The number of slow queries that have taken more than `long_query_time` seconds.  | queries |
| `performance/table_locks_waited` | [table_locks_waited](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Table_locks_waited) | The total number of times that a request for a table lock could not be granted immediately and a wait was needed. | count |
| `performance/threads_connected` | [threads_connected](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Threads_connected) | The number of currently open connections. | connections |
| `performance/threads_running` | [threads_running](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html#statvar_Threads_running) | The number of threads that are not sleeping. | threads |
| `variables/max_connections` | [max_connections](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_max_connections) | The maximum permitted number of simultaneous client connections. | connections |
| `variables/open_files_limit` | [open_files_limit](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_open_files_limit) |  The number of files that the operating system permits [ **mysqld** ](https://dev.mysql.com/doc/refman/5.6/en/mysqld.html "4.3.1 mysqld — The MySQL Server") to open. | files |
| `variables/read_only` | [read_only](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_read_only) | Whether the server is in read-only mode | boolean |

<a name='system-metrics'>

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

<a name='galera-metrics'>

## Galera Metrics
Useful when deploying [cf-mysql-release](https://github.com/cloudfoundry/cf-mysql-release) or [pxc-release](https://github.com/cloudfoundry-incubator/pxc-release) in a galera topology

Metric Name | Galera Status Name | Description | Units |
|------------|---------|-----------------------|-------------------------- |
| `galera/wsrep_ready` | [wsrep_ready](http://galeracluster.com/library/documentation/galera-status-variables.html#wsrep-ready) | Shows whether the node can accept queries. | boolean |
| `galera/wsrep_cluster_size` | [wsrep_cluster_size](http://galeracluster.com/library/documentation/galera-status-variables.html#wsrep-cluster-size) | The current number of nodes in the Galera cluster. | nodes |
| `galera/wsrep_cluster_status` | [wsrep_cluster_status](http://galeracluster.com/library/documentation/galera-status-variables.html#wsrep-cluster-status) | Shows the status of the cluster component, which is whether the node is `PRIMARY` or `NON_PRIMARY`. | State ID.<br /> Values are `PRIMARY`, `NON-PRIMARY`, or `DISCONNECTED`  |
| `galera/wsrep_flow_control_paused` | [wsrep_flow_control_paused](http://galeracluster.com/library/documentation/galera-status-variables.html#wsrep-flow-control-paused) | The fraction of time that replication was paused due to flow control since the server started or last `FLUSH STATUS`. This is a measure of how much replication lag is slowing down the cluster. | float |
| `galera/wsrep_flow_control_sent` | [wsrep_flow_control_sent](http://galeracluster.com/library/documentation/galera-status-variables.html#wsrep-flow-control-sent) | Number of `FC_PAUSE` (flow control pause) events sent by this node. Unlike many status variables, the counter for this one does not reset every time you run the query. | count |
| `galera/wsrep_flow_control_recv` | [wsrep_flow_control_recv](http://galeracluster.com/library/documentation/galera-status-variables.html#wsrep-flow-control-recv) | Number of `FC_PAUSE` (flow control pause) events received by this node. This includes `FC_PAUSE` events sent by this node (it receives from itself). Unlike most status variables, the counter for this one does not reset every time you run the query. | count |
| `galera/wsrep_local_recv_queue` | [wsrep_local_recv_queue](http://galeracluster.com/library/documentation/galera-status-variables.html#wsrep-local-recv-queue) | The instantaneous size of the local received queue. | float |
| `galera/wsrep_local_send_queue` | [wsrep_local_send_queue](http://galeracluster.com/library/documentation/galera-status-variables.html#wsrep-local-send-queue) | The instantaneous size of the local sent queue. | float |
| `galera/wsrep_local_index` | [wsrep_local_index](http://galeracluster.com/library/documentation/galera-status-variables.html#wsrep-local-index) | This node's index in the cluster (base 0). | int | 
| `galera/wsrep_local_state` | [wsrep_local_state](http://galeracluster.com/library/documentation/galera-status-variables.html#wsrep-local-state) | This is the node's local state. | int<br>1 = `JOINING`<br>2 = `DONOR/DESYNCED`<br>3 = `JOINED`<br>4 = `SYNCED`|

<a name='leader-follower-metrics'>

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

<a name='broker-metrics'>

## Broker Metrics
Can be implemented when deploying mysql releases with a service broker

Metric Name | Description | Units |
|------------|--------------------------------|-------------------------- |
| `broker/disk_allocated_service_plans` | The number of MB allocated by the broker for all service plans, current and allocated. | MB |
