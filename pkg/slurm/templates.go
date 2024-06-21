package slurm

var (
	slurmConfTpl = `
{{ $slikName := .SlikName }}
ClusterName=cluster
SlurmctldHost={{ $slikName }}-slurmctld
ProctrackType=proctrack/linuxproc
ReturnToService=2
SlurmctldPidFile=/run/slurmctld.pid
SlurmdPidFile=/run/slurmd.pid
SlurmdSpoolDir=/var/lib/slurm/slurmd
StateSaveLocation=/var/lib/slurm/slurmctld
SlurmUser=root
TaskPlugin=task/none
SchedulerType=sched/backfill
SelectType=select/cons_tres
SelectTypeParameters=CR_Core_Memory
JobCompType=jobcomp/none
JobAcctGatherType=jobacct_gather/none
SlurmctldDebug=verbose
SlurmctldLogFile=/var/log/slurm/slurmctld.log
SlurmdDebug=verbose
SlurmdLogFile=/var/log/slurm/slurmd.log

# slurmdbd
{{ if .Slurmdbd -}}
AccountingStorageType=accounting_storage/slurmdbd
AccountingStoragePort=6819
AccountingStorageHost={{ $slikName }}-slurmdbd
{{ else }}
AccountingStorageType=accounting_storage/none
{{ end }}

# nodes
NodeName=DEFAULT State=UNKNOWN CPUs=1 CoresPerSocket=1 ThreadsPerCore=1
{{ range .SlurmdNodes -}}
NodeName={{ $slikName }}-{{ .NodeName }} CPUs={{ .CPUs }} RealMemory={{ .RealMemory }} ThreadsPerCore={{ .ThreadsPerCore }}
{{ end }}

# TODO other?
PartitionName=DEFAULT Nodes=ALL MaxTime=60 State=UP
PartitionName=batch Nodes=ALL Default=YES MaxTime=60 State=Up
#PartitionName=debug Nodes=ALL Default=YES MaxTime=INFINITE State=UP
`

	slurmdbdConfTpl = `
{{ $slikName := .SlikName }}
AuthType=auth/munge

DbdHost={{ $slikName }}-slurmdbd
DbdPort=6819

DebugLevel=verbose
MessageTimeout=10

StorageHost={{ $slikName }}-mariadb
StorageLoc=slurmdbd
StorageUser={{ .User }}
StoragePass={{ .Pass }}
StoragePort=3306

StorageType=accounting_storage/mysql

LogFile=/var/log/slurm/slurmdbd.log
PidFile=/run/slurmdbd.pid
SlurmUser=root
`

	slurmInit = `
GRANT ALL PRIVILEGES ON *.* TO 'slurm'@'%';
FLUSH PRIVILEGES;
`

	overridesCnf = `
[mysqld]
max_connections         = 100
table_definition_cache  = 2000

# Per connection settings
max_binlog_cache_size      = "256M"
max_binlog_stmt_cache_size = "64M"

# InnoDB
default_storage_engine          = InnoDB
innodb_buffer_pool_size         = "128M"
innodb_log_buffer_size          = "128M"
innodb_log_file_size            = "256M"
innodb_flush_method             = O_DIRECT
innodb_lock_wait_timeout        = 900

# Logging
general_log                   = 1
general_log_file              = mysql.log
log_error                     = error.log
log_warnings                  = 9
slow_query_log                = 1
slow_query_log_file           = slow.log
long_query_time               = 5
log_slow_rate_limit           = 1
log_slow_verbosity            = "query_plan,explain"
log_queries_not_using_indexes = 0
`
)
