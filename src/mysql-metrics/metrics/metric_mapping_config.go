package metrics

type MetricMappingConfig struct {
	MysqlMetricMappings          map[string]MetricDefinition
	GaleraMetricMappings         map[string]MetricDefinition
	LeaderFollowerMetricMappings map[string]MetricDefinition
	DiskMetricMappings           map[string]MetricDefinition
	BrokerMetricMappings         map[string]MetricDefinition
	CPUMetricMappings            map[string]MetricDefinition
}

type MetricDefinition struct {
	Key  string `yaml:"key"`
	Unit string `yaml:"unit"`
}

func DefaultMetricMappingConfig() *MetricMappingConfig {
	return &MetricMappingConfig{
		MysqlMetricMappings: map[string]MetricDefinition{
			"available": {
				Key:  "available",
				Unit: "boolean",
			},
			"max_connections": {
				Key:  "variables/max_connections",
				Unit: "integer",
			},
			"open_files_limit": {
				Key:  "variables/open_files_limit",
				Unit: "integer",
			},
			"read_only": {
				Key:  "variables/read_only",
				Unit: "boolean",
			},
			"questions": {
				Key:  "performance/questions",
				Unit: "metric",
			},
			"queries": {
				Key:  "performance/queries",
				Unit: "metric",
			},
			"queries_delta": {
				Key:  "performance/queries_delta",
				Unit: "metric",
			},
			"innodb_buffer_pool_pages_free": {
				Key:  "innodb/buffer_pool_pages_free",
				Unit: "page",
			},
			"innodb_buffer_pool_pages_total": {
				Key:  "innodb/buffer_pool_pages_total",
				Unit: "page",
			},
			"innodb_buffer_pool_pages_data": {
				Key:  "innodb/buffer_pool_pages_data",
				Unit: "page",
			},
			"innodb_row_lock_current_waits": {
				Key:  "innodb/row_lock_current_waits",
				Unit: "lock",
			},
			"innodb_data_read": {
				Key:  "innodb/data_read",
				Unit: "byte",
			},
			"innodb_data_written": {
				Key:  "innodb/data_written",
				Unit: "byte",
			},
			"innodb_mutex_os_waits": {
				Key:  "innodb/mutex_os_waits",
				Unit: "event",
			},
			"innodb_mutex_spin_rounds": {
				Key:  "innodb/mutex_spin_rounds",
				Unit: "event",
			},
			"innodb_mutex_spin_waits": {
				Key:  "innodb/mutex_spin_waits",
				Unit: "event",
			},
			"innodb_os_log_fsyncs": {
				Key:  "innodb/os_log_fsyncs",
				Unit: "event",
			},
			"innodb_row_lock_time": {
				Key:  "innodb/row_lock_time",
				Unit: "millisecond",
			},
			"innodb_row_lock_waits": {
				Key:  "innodb/row_lock_waits",
				Unit: "event",
			},
			"connections": {
				Key:  "net/connections",
				Unit: "connection",
			},
			"max_used_connections": {
				Key:  "net/max_used_connections",
				Unit: "connection",
			},
			"com_delete": {
				Key:  "performance/com_delete",
				Unit: "query",
			},
			"com_delete_multi": {
				Key:  "performance/com_delete_multi",
				Unit: "query",
			},
			"com_insert": {
				Key:  "performance/com_insert",
				Unit: "query",
			},
			"com_insert_select": {
				Key:  "performance/com_insert_select",
				Unit: "query",
			},
			"com_replace_select": {
				Key:  "performance/com_replace_select",
				Unit: "query",
			},
			"com_select": {
				Key:  "performance/com_select",
				Unit: "query",
			},
			"com_update": {
				Key:  "performance/com_update",
				Unit: "query",
			},
			"com_update_multi": {
				Key:  "performance/com_update_multi",
				Unit: "query",
			},
			"created_tmp_disk_tables": {
				Key:  "performance/created_tmp_disk_tables",
				Unit: "table",
			},
			"created_tmp_files": {
				Key:  "performance/created_tmp_files",
				Unit: "file",
			},
			"created_tmp_tables": {
				Key:  "performance/created_tmp_tables",
				Unit: "table",
			},
			"cpu_time": {
				Key:  "performance/cpu_time",
				Unit: "second",
			},
			"open_files": {
				Key:  "performance/open_files",
				Unit: "file",
			},
			"open_tables": {
				Key:  "performance/open_tables",
				Unit: "integer",
			},
			"opened_tables": {
				Key:  "performance/opened_tables",
				Unit: "integer",
			},
			"open_table_definitions": {
				Key:  "performance/open_table_definitions",
				Unit: "integer",
			},
			"opened_table_definitions": {
				Key:  "performance/opened_table_definitions",
				Unit: "integer",
			},
			"qcache_hits": {
				Key:  "performance/qcache_hits",
				Unit: "hit",
			},
			"slow_queries": {
				Key:  "performance/slow_queries",
				Unit: "query",
			},
			"table_locks_waited": {
				Key:  "performance/table_locks_waited",
				Unit: "number",
			},
			"threads_connected": {
				Key:  "performance/threads_connected",
				Unit: "connection",
			},
			"threads_running": {
				Key:  "performance/threads_running",
				Unit: "thread",
			},
			"rpl_semi_sync_master_tx_avg_wait_time": {
				Key:  "rpl_semi_sync_master_tx_avg_wait_time",
				Unit: "microsecond",
			},
			"rpl_semi_sync_master_no_tx": {
				Key:  "rpl_semi_sync_master_no_tx",
				Unit: "integer",
			},
			"rpl_semi_sync_master_wait_sessions": {
				Key:  "rpl_semi_sync_master_wait_sessions",
				Unit: "integer",
			},
		},
		GaleraMetricMappings: map[string]MetricDefinition{
			"wsrep_cluster_size": {
				Key:  "galera/wsrep_cluster_size",
				Unit: "node",
			},
			"wsrep_local_recv_queue": {
				Key:  "galera/wsrep_local_recv_queue",
				Unit: "float",
			},
			"wsrep_local_send_queue": {
				Key:  "galera/wsrep_local_send_queue",
				Unit: "float",
			},
			"wsrep_local_index": {
				Key:  "galera/wsrep_local_index",
				Unit: "float",
			},
			"wsrep_local_state": {
				Key:  "galera/wsrep_local_state",
				Unit: "float",
			},
			"wsrep_ready": {
				Key:  "galera/wsrep_ready",
				Unit: "number",
			},
			"wsrep_cluster_status": {
				Key:  "galera/wsrep_cluster_status",
				Unit: "number",
			},
		},
		LeaderFollowerMetricMappings: map[string]MetricDefinition{
			"is_follower": {
				Key:  "follower/is_follower",
				Unit: "boolean",
			},
			"seconds_behind_master": {
				Key:  "follower/seconds_behind_master",
				Unit: "integer",
			},
			"seconds_since_leader_heartbeat": {
				Key:  "follower/seconds_since_leader_heartbeat",
				Unit: "integer",
			},
			"relay_log_space": {
				Key:  "follower/relay_log_space",
				Unit: "bytes",
			},
			"slave_io_running": {
				Key:  "follower/slave_io_running",
				Unit: "boolean",
			},
			"slave_sql_running": {
				Key:  "follower/slave_sql_running",
				Unit: "boolean",
			},
		},
		DiskMetricMappings: map[string]MetricDefinition{
			"persistent_disk_used_percent": {
				Key:  "system/persistent_disk_used_percent",
				Unit: "percentage",
			},
			"persistent_disk_used": {
				Key:  "system/persistent_disk_used",
				Unit: "kb",
			},
			"persistent_disk_free": {
				Key:  "system/persistent_disk_free",
				Unit: "kb",
			},
			"persistent_disk_inodes_used_percent": {
				Key:  "system/persistent_disk_inodes_used_percent",
				Unit: "percentage",
			},
			"persistent_disk_inodes_used": {
				Key:  "system/persistent_disk_inodes_used",
				Unit: "kb",
			},
			"persistent_disk_inodes_free": {
				Key:  "system/persistent_disk_inodes_free",
				Unit: "kb",
			},
			"ephemeral_disk_used_percent": {
				Key:  "system/ephemeral_disk_used_percent",
				Unit: "percentage",
			},
			"ephemeral_disk_used": {
				Key:  "system/ephemeral_disk_used",
				Unit: "kb",
			},
			"ephemeral_disk_free": {
				Key:  "system/ephemeral_disk_free",
				Unit: "kb",
			},
			"ephemeral_disk_inodes_used_percent": {
				Key:  "system/ephemeral_disk_inodes_used_percent",
				Unit: "percentage",
			},
			"ephemeral_disk_inodes_used": {
				Key:  "system/ephemeral_disk_inodes_used",
				Unit: "kb",
			},
			"ephemeral_disk_inodes_free": {
				Key:  "system/ephemeral_disk_inodes_free",
				Unit: "kb",
			},
		},
		BrokerMetricMappings: map[string]MetricDefinition{
			"service_plans_disk_allocated": {
				Key:  "broker/disk_allocated_service_plans",
				Unit: "megabyte",
			},
		},
		CPUMetricMappings: map[string]MetricDefinition{
			"cpu_utilization_percent": {
				Key:  "performance/cpu_utilization_percent",
				Unit: "percentage",
			},
		},
	}
}
