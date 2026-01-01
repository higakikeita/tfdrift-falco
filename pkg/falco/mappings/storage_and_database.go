package mappings

// StorageAndDatabaseMappings contains CloudTrail event to Terraform resource mappings for storage and database services
var StorageAndDatabaseMappings = map[string]string{
	// S3 - Bucket Management
	"CreateBucket": "aws_s3_bucket",
	"DeleteBucket": "aws_s3_bucket",

	// S3 - Bucket Configuration
	"PutBucketPolicy":               "aws_s3_bucket_policy",
	"DeleteBucketPolicy":            "aws_s3_bucket_policy",
	"PutBucketVersioning":           "aws_s3_bucket_versioning",
	"PutBucketEncryption":           "aws_s3_bucket",
	"DeleteBucketEncryption":        "aws_s3_bucket",
	"PutBucketLogging":              "aws_s3_bucket_logging",
	"PutBucketPublicAccessBlock":    "aws_s3_bucket_public_access_block",
	"DeleteBucketPublicAccessBlock": "aws_s3_bucket_public_access_block",
	"PutBucketAcl":                  "aws_s3_bucket_acl",

	// S3 - Lifecycle Management
	"PutBucketLifecycle":    "aws_s3_bucket_lifecycle_configuration",
	"DeleteBucketLifecycle": "aws_s3_bucket_lifecycle_configuration",

	// S3 - Replication
	"PutBucketReplication":    "aws_s3_bucket_replication_configuration",
	"DeleteBucketReplication": "aws_s3_bucket_replication_configuration",

	// S3 - CORS
	"PutBucketCors":    "aws_s3_bucket_cors_configuration",
	"DeleteBucketCors": "aws_s3_bucket_cors_configuration",

	// S3 - Website
	"PutBucketWebsite":    "aws_s3_bucket_website_configuration",
	"DeleteBucketWebsite": "aws_s3_bucket_website_configuration",

	// S3 - Tagging
	"PutBucketTagging":    "aws_s3_bucket",
	"DeleteBucketTagging": "aws_s3_bucket",

	// S3 - Other
	"PutBucketRequestPayment":   "aws_s3_bucket_request_payment_configuration",
	"PutBucketAccelerateConfig": "aws_s3_bucket_accelerate_configuration",

	// RDS - DB Instances
	"CreateDBInstance": "aws_db_instance",
	"DeleteDBInstance": "aws_db_instance",
	"ModifyDBInstance": "aws_db_instance",
	"RebootDBInstance": "aws_db_instance",
	"StartDBInstance":  "aws_db_instance",
	"StopDBInstance":   "aws_db_instance",

	// RDS - DB Clusters (Aurora)
	"CreateDBCluster":   "aws_rds_cluster",
	"DeleteDBCluster":   "aws_rds_cluster",
	"ModifyDBCluster":   "aws_rds_cluster",
	"StartDBCluster":    "aws_rds_cluster",
	"StopDBCluster":     "aws_rds_cluster",
	"FailoverDBCluster": "aws_rds_cluster",

	// RDS - Aurora Specific
	"AddRoleToDBCluster":      "aws_rds_cluster_role_association",
	"RemoveRoleFromDBCluster": "aws_rds_cluster_role_association",
	"ModifyDBClusterEndpoint": "aws_rds_cluster_endpoint",
	"CreateDBClusterEndpoint": "aws_rds_cluster_endpoint",
	"DeleteDBClusterEndpoint": "aws_rds_cluster_endpoint",
	"ModifyGlobalCluster":     "aws_rds_global_cluster",

	// RDS - Snapshots
	"CreateDBSnapshot":          "aws_db_snapshot",
	"DeleteDBSnapshot":          "aws_db_snapshot",
	"ModifyDBSnapshotAttribute": "aws_db_snapshot",
	"CreateDBClusterSnapshot":   "aws_db_cluster_snapshot",
	"DeleteDBClusterSnapshot":   "aws_db_cluster_snapshot",

	// RDS - Parameter Groups
	"CreateDBParameterGroup": "aws_db_parameter_group",
	"DeleteDBParameterGroup": "aws_db_parameter_group",
	"ModifyDBParameterGroup": "aws_db_parameter_group",

	// RDS - Subnet Groups
	"CreateDBSubnetGroup": "aws_db_subnet_group",
	"DeleteDBSubnetGroup": "aws_db_subnet_group",
	"ModifyDBSubnetGroup": "aws_db_subnet_group",

	// RDS - Option Groups
	"CreateOptionGroup": "aws_db_option_group",
	"DeleteOptionGroup": "aws_db_option_group",
	"ModifyOptionGroup": "aws_db_option_group",

	// RDS - Security & Backup
	"ModifyDBInstanceAttribute":       "aws_db_instance",
	"RestoreDBInstanceFromDBSnapshot": "aws_db_instance",
	"RestoreDBInstanceToPointInTime":  "aws_db_instance",
	"RestoreDBClusterFromSnapshot":    "aws_rds_cluster",
	"CreateDBInstanceReadReplica":     "aws_db_instance",

	// DynamoDB - Tables
	"CreateTable":             "aws_dynamodb_table",
	"DeleteTable":             "aws_dynamodb_table",
	"UpdateTable":             "aws_dynamodb_table",
	"UpdateTimeToLive":        "aws_dynamodb_table",
	"UpdateContinuousBackups": "aws_dynamodb_table",

	// DynamoDB - Global Tables
	"CreateGlobalTable": "aws_dynamodb_global_table",
	"UpdateGlobalTable": "aws_dynamodb_global_table",

	// DynamoDB - Backups
	"CreateBackup":              "aws_dynamodb_table_backup",
	"DeleteBackup":              "aws_dynamodb_table_backup",
	"RestoreTableFromBackup":    "aws_dynamodb_table",
	"RestoreTableToPointInTime": "aws_dynamodb_table",

	// DynamoDB - Streams
	"EnableKinesisStreamingDestination":  "aws_dynamodb_kinesis_streaming_destination",
	"DisableKinesisStreamingDestination": "aws_dynamodb_kinesis_streaming_destination",

	// DynamoDB - Contributor Insights
	"UpdateContributorInsights": "aws_dynamodb_contributor_insights",

	// ElastiCache - Cache Clusters
	"CreateCacheCluster": "aws_elasticache_cluster",
	"DeleteCacheCluster": "aws_elasticache_cluster",
	"ModifyCacheCluster": "aws_elasticache_cluster",
	"RebootCacheCluster": "aws_elasticache_cluster",

	// ElastiCache - Replication Groups
	"CreateReplicationGroup": "aws_elasticache_replication_group",
	"DeleteReplicationGroup": "aws_elasticache_replication_group",
	"ModifyReplicationGroup": "aws_elasticache_replication_group",

	// ElastiCache - Parameter Groups
	"CreateCacheParameterGroup": "aws_elasticache_parameter_group",
	"DeleteCacheParameterGroup": "aws_elasticache_parameter_group",
	"ModifyCacheParameterGroup": "aws_elasticache_parameter_group",

	// ElastiCache - Subnet Groups
	"CreateCacheSubnetGroup": "aws_elasticache_subnet_group",
	"DeleteCacheSubnetGroup": "aws_elasticache_subnet_group",
	"ModifyCacheSubnetGroup": "aws_elasticache_subnet_group",

	// ElastiCache - Replica Management
	"IncreaseReplicaCount": "aws_elasticache_replication_group",
	"DecreaseReplicaCount": "aws_elasticache_replication_group",

	// ElastiCache - Global Replication
	"CreateGlobalReplicationGroup": "aws_elasticache_global_replication_group",
	"DeleteGlobalReplicationGroup": "aws_elasticache_global_replication_group",

	// Redshift - Clusters
	"CreateCluster": "aws_redshift_cluster",
	"DeleteCluster": "aws_redshift_cluster",
	"ModifyCluster": "aws_redshift_cluster",
	"RebootCluster": "aws_redshift_cluster",
	"ResizeCluster": "aws_redshift_cluster",

	// ECR - Repositories
	"CreateRepository": "aws_ecr_repository",
	"DeleteRepository": "aws_ecr_repository",

	// ECR - Configuration
	"PutImageScanningConfiguration": "aws_ecr_repository",
	"PutImageTagMutability":         "aws_ecr_repository",
	"PutLifecyclePolicy":            "aws_ecr_lifecycle_policy",
	"DeleteLifecyclePolicy":         "aws_ecr_lifecycle_policy",
	"SetRepositoryPolicy":           "aws_ecr_repository_policy",
	"DeleteRepositoryPolicy":        "aws_ecr_repository_policy",
	"PutReplicationConfiguration":   "aws_ecr_replication_configuration",
}
