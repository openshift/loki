package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	configv1 "github.com/grafana/loki/operator/api/config/v1"
	lokiv1 "github.com/grafana/loki/operator/api/loki/v1"
	"github.com/grafana/loki/operator/internal/manifests/storage"
)

func TestBuild_ConfigAndRuntimeConfig_NoRuntimeConfigGenerated(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/no-runtime-config/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 0,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		DiscoverLogLevels:     true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_BothGenerated(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/both-generated/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
  test-a:
    ingestion_rate_mb: 2
    ingestion_burst_size_mb: 5
    max_global_streams_per_user: 1
    max_chunks_per_query: 1000000
    blocked_queries:
    - pattern: ""
      hash: 12345
      types: metric,limited
    - pattern: .*prod.*
      regex: true
    - pattern: ""
      types: metric
    - pattern: sum(rate({env="prod"}[1m]))
    - pattern: '{kubernetes_namespace_name="my-app"}'
    - pattern: ""
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
				Tenants: map[string]lokiv1.PerTenantLimitsTemplateSpec{
					"test-a": {
						IngestionLimits: &lokiv1.IngestionLimitSpec{
							IngestionRate:             2,
							IngestionBurstSize:        5,
							MaxGlobalStreamsPerTenant: 1,
						},
						QueryLimits: &lokiv1.PerTenantQueryLimitSpec{
							QueryLimitSpec: lokiv1.QueryLimitSpec{
								MaxChunksPerQuery: 1000000,
							},
							Blocked: []lokiv1.BlockedQuerySpec{
								{
									Hash:  12345,
									Types: lokiv1.BlockedQueryTypes{lokiv1.BlockedQueryMetric, lokiv1.BlockedQueryLimited},
								},
								{
									Pattern: ".*prod.*",
									Regex:   true,
								},
								{
									Types: lokiv1.BlockedQueryTypes{lokiv1.BlockedQueryMetric},
								},
								{
									Pattern: `sum(rate({env="prod"}[1m]))`,
								},
								{
									Pattern: `{kubernetes_namespace_name="my-app"}`,
								},
								{
									Pattern: "",
								},
							},
						},
					},
				},
			},
		},
		Overrides: map[string]LokiOverrides{
			"test-a": {
				Limits: lokiv1.PerTenantLimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             2,
						MaxGlobalStreamsPerTenant: 1,
						IngestionBurstSize:        5,
					},
					QueryLimits: &lokiv1.PerTenantQueryLimitSpec{
						QueryLimitSpec: lokiv1.QueryLimitSpec{
							MaxChunksPerQuery: 1000000,
						},
						Blocked: []lokiv1.BlockedQuerySpec{
							{
								Hash:  12345,
								Types: lokiv1.BlockedQueryTypes{lokiv1.BlockedQueryMetric, lokiv1.BlockedQueryLimited},
							},
							{
								Pattern: ".*prod.*",
								Regex:   true,
							},
							{
								Types: lokiv1.BlockedQueryTypes{lokiv1.BlockedQueryMetric},
							},
							{
								Pattern: `sum(rate({env="prod"}[1m]))`,
							},
							{
								Pattern: `{kubernetes_namespace_name="my-app"}`,
							},
							{
								Pattern: "",
							},
						},
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers: []string{"boltdb"},
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_CreateLokiConfigFailed(t *testing.T) {
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					// making it nil so that the template is not generated and error is returned
					QueryLimits: nil,
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers: []string{"boltdb"},
	}
	cfg, rCfg, err := Build(opts)
	require.Error(t, err)
	require.Empty(t, cfg)
	require.Empty(t, rCfg)
}

func TestBuild_ConfigAndRuntimeConfig_RulerConfigGenerated_WithHeaderAuthorization(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/ruler-with-auth-header/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		Ruler: Ruler{
			Enabled:               true,
			RulesStorageDirectory: "/tmp/rules",
			EvaluationInterval:    "1m",
			PollInterval:          "1m",
			AlertManager: &AlertManagerConfig{
				ExternalURL: "http://alert.me/now",
				ExternalLabels: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
				Hosts:              "http://alerthost1,http://alerthost2",
				EnableV2:           true,
				EnableDiscovery:    true,
				RefreshInterval:    "1m",
				QueueCapacity:      1000,
				Timeout:            "1m",
				ForOutageTolerance: "10m",
				ForGracePeriod:     "5m",
				ResendDelay:        "2m",
			},
			RemoteWrite: &RemoteWriteConfig{
				Enabled:       true,
				RefreshPeriod: "1m",
				Client: &RemoteWriteClientConfig{
					Name:          "remote-write-me",
					URL:           "http://remote.write.me",
					RemoteTimeout: "10s",
					Headers: map[string]string{
						"more": "foryou",
						"less": "forme",
					},
					ProxyURL:        "http://proxy.through.me",
					FollowRedirects: true,
					BearerToken:     "supersecret",
				},
				Queue: &RemoteWriteQueueConfig{
					Capacity:          1000,
					MaxShards:         100,
					MinShards:         50,
					MaxSamplesPerSend: 1000,
					BatchSendDeadline: "10s",
					MinBackOffPeriod:  "30ms",
					MaxBackOffPeriod:  "100ms",
				},
			},
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_RulerConfigGenerated_WithBasicAuthorization(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/ruler-with-auth-basic/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		Ruler: Ruler{
			Enabled:               true,
			RulesStorageDirectory: "/tmp/rules",
			EvaluationInterval:    "1m",
			PollInterval:          "1m",
			AlertManager: &AlertManagerConfig{
				ExternalURL: "http://alert.me/now",
				ExternalLabels: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
				Hosts:              "http://alerthost1,http://alerthost2",
				EnableV2:           true,
				EnableDiscovery:    true,
				RefreshInterval:    "1m",
				QueueCapacity:      1000,
				Timeout:            "1m",
				ForOutageTolerance: "10m",
				ForGracePeriod:     "5m",
				ResendDelay:        "2m",
			},
			RemoteWrite: &RemoteWriteConfig{
				Enabled:       true,
				RefreshPeriod: "1m",
				Client: &RemoteWriteClientConfig{
					Name:          "remote-write-me",
					URL:           "http://remote.write.me",
					RemoteTimeout: "10s",
					Headers: map[string]string{
						"more": "foryou",
						"less": "forme",
					},
					ProxyURL:          "http://proxy.through.me",
					FollowRedirects:   true,
					BasicAuthUsername: "user",
					BasicAuthPassword: "passwd",
				},
				Queue: &RemoteWriteQueueConfig{
					Capacity:          1000,
					MaxShards:         100,
					MinShards:         50,
					MaxSamplesPerSend: 1000,
					BatchSendDeadline: "10s",
					MinBackOffPeriod:  "30ms",
					MaxBackOffPeriod:  "100ms",
				},
			},
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_RulerConfigGenerated_WithRelabelConfigs(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/ruler-with-relabel-configs/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		Ruler: Ruler{
			Enabled:               true,
			RulesStorageDirectory: "/tmp/rules",
			EvaluationInterval:    "1m",
			PollInterval:          "1m",
			AlertManager: &AlertManagerConfig{
				ExternalURL: "http://alert.me/now",
				ExternalLabels: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
				Hosts:              "http://alerthost1,http://alerthost2",
				EnableV2:           true,
				EnableDiscovery:    true,
				RefreshInterval:    "1m",
				QueueCapacity:      1000,
				Timeout:            "1m",
				ForOutageTolerance: "10m",
				ForGracePeriod:     "5m",
				ResendDelay:        "2m",
			},
			RemoteWrite: &RemoteWriteConfig{
				Enabled:       true,
				RefreshPeriod: "1m",
				Client: &RemoteWriteClientConfig{
					Name:          "remote-write-me",
					URL:           "http://remote.write.me",
					RemoteTimeout: "10s",
					Headers: map[string]string{
						"more": "foryou",
						"less": "forme",
					},
					ProxyURL:          "http://proxy.through.me",
					FollowRedirects:   true,
					BasicAuthUsername: "user",
					BasicAuthPassword: "passwd",
				},
				RelabelConfigs: []RelabelConfig{
					{
						SourceLabels: []string{"labela", "labelb"},
						Regex:        "ALERTS.*",
						Action:       "drop",
						Separator:    "\\",
						Replacement:  "$1",
					},
					{
						SourceLabels: []string{"labelc", "labeld"},
						Regex:        "ALERTS.*",
						Action:       "drop",
						Replacement:  "$1",
						TargetLabel:  "labeld",
						Modulus:      123,
					},
				},
				Queue: &RemoteWriteQueueConfig{
					Capacity:          1000,
					MaxShards:         100,
					MinShards:         50,
					MaxSamplesPerSend: 1000,
					BatchSendDeadline: "10s",
					MinBackOffPeriod:  "30ms",
					MaxBackOffPeriod:  "100ms",
				},
			},
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_WithRetention(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/with-retention/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
  test-a:
    ingestion_rate_mb: 2
    ingestion_burst_size_mb: 5
    max_global_streams_per_user: 1
    max_chunks_per_query: 1000000
    retention_period: 7d
    retention_stream:
    - period: 15d
      priority: 1
      selector: "{environment=\"production\"}"
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
					Retention: &lokiv1.RetentionLimitSpec{
						Days: 15,
						Streams: []*lokiv1.RetentionStreamSpec{
							{
								Days:     3,
								Priority: 1,
								Selector: `{environment="development"}`,
							},
						},
					},
				},
				Tenants: map[string]lokiv1.PerTenantLimitsTemplateSpec{
					"test-a": {
						IngestionLimits: &lokiv1.IngestionLimitSpec{
							IngestionRate:             2,
							IngestionBurstSize:        5,
							MaxGlobalStreamsPerTenant: 1,
						},
						QueryLimits: &lokiv1.PerTenantQueryLimitSpec{
							QueryLimitSpec: lokiv1.QueryLimitSpec{
								MaxChunksPerQuery: 1000000,
							},
						},
						Retention: &lokiv1.RetentionLimitSpec{
							Days: 7,
							Streams: []*lokiv1.RetentionStreamSpec{
								{
									Days:     15,
									Priority: 1,
									Selector: `{environment="production"}`,
								},
							},
						},
					},
				},
			},
		},
		Overrides: map[string]LokiOverrides{
			"test-a": {
				Limits: lokiv1.PerTenantLimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             2,
						IngestionBurstSize:        5,
						MaxGlobalStreamsPerTenant: 1,
					},
					QueryLimits: &lokiv1.PerTenantQueryLimitSpec{
						QueryLimitSpec: lokiv1.QueryLimitSpec{
							MaxChunksPerQuery: 1000000,
						},
					},
					Retention: &lokiv1.RetentionLimitSpec{
						Days: 7,
						Streams: []*lokiv1.RetentionStreamSpec{
							{
								Days:     15,
								Priority: 1,
								Selector: `{environment="production"}`,
							},
						},
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers: []string{"boltdb"},
		Retention: RetentionOptions{
			Enabled:           true,
			DeleteWorkerCount: 50,
		},
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_RulerConfigGenerated_WithAlertRelabelConfigs(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/ruler-with-alert-relabel-configs/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "2m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		Ruler: Ruler{
			Enabled:               true,
			RulesStorageDirectory: "/tmp/rules",
			EvaluationInterval:    "1m",
			PollInterval:          "1m",
			AlertManager: &AlertManagerConfig{
				ExternalURL: "http://alert.me/now",
				ExternalLabels: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
				Hosts:              "http://alerthost1,http://alerthost2",
				EnableV2:           true,
				EnableDiscovery:    true,
				RefreshInterval:    "1m",
				QueueCapacity:      1000,
				Timeout:            "1m",
				ForOutageTolerance: "10m",
				ForGracePeriod:     "5m",
				ResendDelay:        "2m",
				RelabelConfigs: []RelabelConfig{
					{
						SourceLabels: []string{"source1", "source2"},
						Regex:        "ALERTS.*",
						Action:       "drop",
						Separator:    "\\",
						Replacement:  "$1",
					},
					{
						SourceLabels: []string{"source3", "source4"},
						Regex:        "ALERTS.*",
						Action:       "keep",
						Replacement:  "$1",
						TargetLabel:  "target",
						Modulus:      42,
					},
				},
			},
			RemoteWrite: &RemoteWriteConfig{
				Enabled:       true,
				RefreshPeriod: "1m",
				Client: &RemoteWriteClientConfig{
					Name:          "remote-write-me",
					URL:           "http://remote.write.me",
					RemoteTimeout: "10s",
					Headers: map[string]string{
						"more": "foryou",
						"less": "forme",
					},
					ProxyURL:          "http://proxy.through.me",
					FollowRedirects:   true,
					BasicAuthUsername: "user",
					BasicAuthPassword: "passwd",
				},
				RelabelConfigs: []RelabelConfig{
					{
						SourceLabels: []string{"labela", "labelb"},
						Regex:        "ALERTS.*",
						Action:       "drop",
						Separator:    "\\",
						Replacement:  "$1",
					},
					{
						SourceLabels: []string{"labelc", "labeld"},
						Regex:        "ALERTS.*",
						Action:       "drop",
						Replacement:  "$1",
						TargetLabel:  "labeld",
						Modulus:      123,
					},
				},
				Queue: &RemoteWriteQueueConfig{
					Capacity:          1000,
					MaxShards:         100,
					MinShards:         50,
					MaxSamplesPerSend: 1000,
					BatchSendDeadline: "10s",
					MinBackOffPeriod:  "30ms",
					MaxBackOffPeriod:  "100ms",
				},
			},
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_WithTLS(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/with-tls/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Gates: configv1.FeatureGates{
			HTTPEncryption: true,
			GRPCEncryption: true,
		},
		TLS: TLSOptions{
			Ciphers:       []string{"cipher1", "cipher2"},
			MinTLSVersion: "VersionTLS12",
			Paths: TLSFilePaths{
				CA: "/var/run/tls/ca.pem",
				GRPC: TLSCertPath{
					Certificate: "/var/run/tls/grpc/tls.crt",
					Key:         "/var/run/tls/grpc/tls.key",
				},
				HTTP: TLSCertPath{
					Certificate: "/var/run/tls/http/tls.crt",
					Key:         "/var/run/tls/http/tls.key",
				},
			},
			ServerNames: TLSServerNames{
				GRPC: GRPCServerNames{
					Compactor:     "compactor-grpc.svc",
					IndexGateway:  "index-gateway-grpc.svc",
					Ingester:      "ingester-grpc.svc",
					QueryFrontend: "query-frontend-grpc.svc",
					Ruler:         "ruler-grpc.svc",
				},
				HTTP: HTTPServerNames{
					Querier: "querier-http.svc",
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint: "http://s3.us-east.amazonaws.com",
				Region:   "us-east",
				Buckets:  "loki",
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_RulerConfigGenerated_WithAlertmanagerOverrides(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/ruler-with-alertmanager-overrides/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
  test-a:
    ruler_alertmanager_config:
      external_url: http://alert.you/later
      external_labels:
        specialkey: specialvalue
      alertmanager_url: http://specialhost
      enable_alertmanager_v2: true
      enable_alertmanager_discovery: true
      alertmanager_refresh_interval: 2m
      notification_queue_capacity: 2000
      notification_timeout: 2m
      alertmanager_client:
        tls_cert_path: "custom/path"
        tls_key_path: "custom/key"
        tls_ca_path: "custom/CA"
        tls_server_name: "custom-servername"
        tls_insecure_skip_verify: false
        basic_auth_password: "pass"
        basic_auth_username: "user"
        credentials: "creds"
        credentials_file: "cred/file"
        type: "auth"
      alert_relabel_configs:
      - source_labels: ["specialsource"]
        regex: "ALERTS.*"
        action: "drop"
        separator: "\\"
        replacement: "$2"
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "2m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		Overrides: map[string]LokiOverrides{
			"test-a": {
				Ruler: RulerOverrides{
					AlertManager: &AlertManagerConfig{
						Hosts:       "http://specialhost",
						ExternalURL: "http://alert.you/later",
						ExternalLabels: map[string]string{
							"specialkey": "specialvalue",
						},
						EnableV2:           true,
						EnableDiscovery:    true,
						RefreshInterval:    "2m",
						QueueCapacity:      2000,
						Timeout:            "2m",
						ForOutageTolerance: "20m",
						ForGracePeriod:     "10m",
						ResendDelay:        "4m",
						RelabelConfigs: []RelabelConfig{
							{
								SourceLabels: []string{"specialsource"},
								Regex:        "ALERTS.*",
								Action:       "drop",
								Separator:    "\\",
								Replacement:  "$2",
							},
						},
						Notifier: &NotifierConfig{
							TLS: TLSConfig{
								ServerName:         ptr.To("custom-servername"),
								CertPath:           ptr.To("custom/path"),
								KeyPath:            ptr.To("custom/key"),
								CAPath:             ptr.To("custom/CA"),
								InsecureSkipVerify: ptr.To(false),
							},
							BasicAuth: BasicAuth{
								Username: ptr.To("user"),
								Password: ptr.To("pass"),
							},
							HeaderAuth: HeaderAuth{
								CredentialsFile: ptr.To("cred/file"),
								Type:            ptr.To("auth"),
								Credentials:     ptr.To("creds"),
							},
						},
					},
				},
			},
		},
		Ruler: Ruler{
			Enabled:               true,
			RulesStorageDirectory: "/tmp/rules",
			EvaluationInterval:    "1m",
			PollInterval:          "1m",

			AlertManager: &AlertManagerConfig{
				ExternalURL: "http://alert.me/now",
				ExternalLabels: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
				Hosts:              "http://alerthost1,http://alerthost2",
				EnableV2:           true,
				EnableDiscovery:    true,
				RefreshInterval:    "1m",
				QueueCapacity:      1000,
				Timeout:            "1m",
				ForOutageTolerance: "10m",
				ForGracePeriod:     "5m",
				ResendDelay:        "2m",
				RelabelConfigs: []RelabelConfig{
					{
						SourceLabels: []string{"source1", "source2"},
						Regex:        "ALERTS.*",
						Action:       "drop",
						Separator:    "\\",
						Replacement:  "$1",
					},
					{
						SourceLabels: []string{"source3", "source4"},
						Regex:        "ALERTS.*",
						Action:       "keep",
						Replacement:  "$1",
						TargetLabel:  "target",
						Modulus:      42,
					},
				},
			},
			RemoteWrite: &RemoteWriteConfig{
				Enabled:       true,
				RefreshPeriod: "1m",
				Client: &RemoteWriteClientConfig{
					Name:          "remote-write-me",
					URL:           "http://remote.write.me",
					RemoteTimeout: "10s",
					Headers: map[string]string{
						"more": "foryou",
						"less": "forme",
					},
					ProxyURL:          "http://proxy.through.me",
					FollowRedirects:   true,
					BasicAuthUsername: "user",
					BasicAuthPassword: "passwd",
				},
				RelabelConfigs: []RelabelConfig{
					{
						SourceLabels: []string{"labela", "labelb"},
						Regex:        "ALERTS.*",
						Action:       "drop",
						Separator:    "\\",
						Replacement:  "$1",
					},
					{
						SourceLabels: []string{"labelc", "labeld"},
						Regex:        "ALERTS.*",
						Action:       "drop",
						Replacement:  "$1",
						TargetLabel:  "labeld",
						Modulus:      123,
					},
				},
				Queue: &RemoteWriteQueueConfig{
					Capacity:          1000,
					MaxShards:         100,
					MinShards:         50,
					MaxSamplesPerSend: 1000,
					BatchSendDeadline: "10s",
					MinBackOffPeriod:  "30ms",
					MaxBackOffPeriod:  "100ms",
				},
			},
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_WithHashRingSpec(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/with-hashring-spec/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstanceAddr:         "${HASH_RING_INSTANCE_ADDR}",
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_WithHashRingSpec_EnableIPv6(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/with-hashring-spec-ipv6/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			EnableIPv6:           true,
			InstanceAddr:         "${HASH_RING_INSTANCE_ADDR}",
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_WithReplicationSpec(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/with-replication-spec/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:                   9095,
			BindPort:                       7946,
			MembersDiscoveryAddr:           "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
			EnableInstanceAvailabilityZone: true,
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_WithS3SSEKMS(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/with-s3-sse-kms/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
  test-a:
    ingestion_rate_mb: 2
    ingestion_burst_size_mb: 5
    max_global_streams_per_user: 1
    max_chunks_per_query: 1000000
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
				Tenants: map[string]lokiv1.PerTenantLimitsTemplateSpec{
					"test-a": {
						IngestionLimits: &lokiv1.IngestionLimitSpec{
							IngestionRate:             2,
							IngestionBurstSize:        5,
							MaxGlobalStreamsPerTenant: 1,
						},
						QueryLimits: &lokiv1.PerTenantQueryLimitSpec{
							QueryLimitSpec: lokiv1.QueryLimitSpec{
								MaxChunksPerQuery: 1000000,
							},
						},
					},
				},
			},
		},
		Overrides: map[string]LokiOverrides{
			"test-a": {
				Limits: lokiv1.PerTenantLimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             2,
						MaxGlobalStreamsPerTenant: 1,
						IngestionBurstSize:        5,
					},
					QueryLimits: &lokiv1.PerTenantQueryLimitSpec{
						QueryLimitSpec: lokiv1.QueryLimitSpec{
							MaxChunksPerQuery: 1000000,
						},
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,

				SSE: storage.S3SSEConfig{
					Type:                 storage.SSEKMSType,
					KMSKeyID:             "test",
					KMSEncryptionContext: `{"key": "value", "another":"value1"}`,
				},
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers: []string{"boltdb"},
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_WithS3SSES3(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/with-s3-sse-s3/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
  test-a:
    ingestion_rate_mb: 2
    ingestion_burst_size_mb: 5
    max_global_streams_per_user: 1
    max_chunks_per_query: 1000000
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
				Tenants: map[string]lokiv1.PerTenantLimitsTemplateSpec{
					"test-a": {
						IngestionLimits: &lokiv1.IngestionLimitSpec{
							IngestionRate:             2,
							IngestionBurstSize:        5,
							MaxGlobalStreamsPerTenant: 1,
						},
						QueryLimits: &lokiv1.PerTenantQueryLimitSpec{
							QueryLimitSpec: lokiv1.QueryLimitSpec{
								MaxChunksPerQuery: 1000000,
							},
						},
					},
				},
			},
		},
		Overrides: map[string]LokiOverrides{
			"test-a": {
				Limits: lokiv1.PerTenantLimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             2,
						MaxGlobalStreamsPerTenant: 1,
						IngestionBurstSize:        5,
					},
					QueryLimits: &lokiv1.PerTenantQueryLimitSpec{
						QueryLimitSpec: lokiv1.QueryLimitSpec{
							MaxChunksPerQuery: 1000000,
						},
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,

				SSE: storage.S3SSEConfig{
					Type:                 storage.SSES3Type,
					KMSKeyID:             "test",
					KMSEncryptionContext: `{"key": "value", "another":"value1"}`,
				},
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers: []string{"boltdb"},
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_WithManualPerStreamRateLimits(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/with-manual-stream-ratelimits/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        3,
						PerStreamRateLimitBurst:   15,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func defaultOptions() Options {
	return Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        3,
						PerStreamRateLimitBurst:   15,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
}

func TestBuild_ConfigAndRuntimeConfig_Schemas(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		schemaConfig            []lokiv1.ObjectStorageSchema
		shippers                []string
		allowStructuredMetadata bool
		expSchemaConfig         string
		expStorageConfig        string
		expStructuredMetadata   string
	}{
		{
			name: "default_config_v11_schema",
			schemaConfig: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
			shippers: []string{"boltdb"},
			expSchemaConfig: `
  configs:
    - from: "2020-10-01"
      index:
        period: 24h
        prefix: index_
      object_store: s3
      schema: v11
      store: boltdb-shipper`,
			expStorageConfig: `
  boltdb_shipper:
    active_index_directory: /tmp/loki/index
    cache_location: /tmp/loki/index_cache
    cache_ttl: 24h
    resync_interval: 5m
    index_gateway_client:
      server_address: dns:///loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local:9095`,
			expStructuredMetadata: `
  allow_structured_metadata: false`,
		},
		{
			name: "v12_schema",
			schemaConfig: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV12,
					EffectiveDate: "2020-02-05",
				},
			},
			shippers: []string{"boltdb"},
			expSchemaConfig: `
  configs:
    - from: "2020-02-05"
      index:
        period: 24h
        prefix: index_
      object_store: s3
      schema: v12
      store: boltdb-shipper`,
			expStorageConfig: `
  boltdb_shipper:
    active_index_directory: /tmp/loki/index
    cache_location: /tmp/loki/index_cache
    cache_ttl: 24h
    resync_interval: 5m
    index_gateway_client:
      server_address: dns:///loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local:9095`,
			expStructuredMetadata: `
  allow_structured_metadata: false`,
		},
		{
			name: "v13_schema",
			schemaConfig: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV13,
					EffectiveDate: "2024-01-01",
				},
			},
			allowStructuredMetadata: true,
			shippers:                []string{"tsdb"},
			expSchemaConfig: `
  configs:
    - from: "2024-01-01"
      index:
        period: 24h
        prefix: index_
      object_store: s3
      schema: v13
      store: tsdb`,
			expStorageConfig: `
  tsdb_shipper:
    active_index_directory: /tmp/loki/tsdb-index
    cache_location: /tmp/loki/tsdb-cache
    cache_ttl: 24h
    resync_interval: 5m
    index_gateway_client:
      server_address: dns:///loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local:9095`,
			expStructuredMetadata: `
  allow_structured_metadata: true`,
		},
		{
			name: "multiple_schema",
			schemaConfig: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-01-01",
				},
				{
					Version:       lokiv1.ObjectStorageSchemaV12,
					EffectiveDate: "2021-01-01",
				},
				{
					Version:       lokiv1.ObjectStorageSchemaV13,
					EffectiveDate: "2024-01-01",
				},
			},
			shippers:                []string{"boltdb", "tsdb"},
			allowStructuredMetadata: true,
			expSchemaConfig: `
  configs:
    - from: "2020-01-01"
      index:
        period: 24h
        prefix: index_
      object_store: s3
      schema: v11
      store: boltdb-shipper
    - from: "2021-01-01"
      index:
        period: 24h
        prefix: index_
      object_store: s3
      schema: v12
      store: boltdb-shipper
    - from: "2024-01-01"
      index:
        period: 24h
        prefix: index_
      object_store: s3
      schema: v13
      store: tsdb`,
			expStorageConfig: `
  boltdb_shipper:
    active_index_directory: /tmp/loki/index
    cache_location: /tmp/loki/index_cache
    cache_ttl: 24h
    resync_interval: 5m
    index_gateway_client:
      server_address: dns:///loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local:9095
  tsdb_shipper:
    active_index_directory: /tmp/loki/tsdb-index
    cache_location: /tmp/loki/tsdb-cache
    cache_ttl: 24h
    resync_interval: 5m
    index_gateway_client:
      server_address: dns:///loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local:9095`,
			expStructuredMetadata: `
  allow_structured_metadata: true`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expCfgBytes, err := os.ReadFile("testdata/schemas-template.yaml")
			require.NoError(t, err)
			expCfg := string(expCfgBytes)
			expCfg = strings.Replace(expCfg, "${SCHEMA_CONFIG}", tc.expSchemaConfig, 1)
			expCfg = strings.Replace(expCfg, "${STORAGE_CONFIG}", tc.expStorageConfig, 1)
			expCfg = strings.Replace(expCfg, "${STORAGE_STRUCTURED_METADATA}", tc.expStructuredMetadata, 1)

			opts := defaultOptions()
			opts.ObjectStorage.Schemas = tc.schemaConfig
			opts.ObjectStorage.AllowStructuredMetadata = tc.allowStructuredMetadata
			opts.Shippers = tc.shippers

			cfg, _, err := Build(opts)
			require.NoError(t, err)
			require.YAMLEq(t, expCfg, string(cfg))
		})
	}
}

func TestBuild_ConfigAndRuntimeConfig_STS(t *testing.T) {
	objStorageConfig := storage.Options{
		SharedStore: lokiv1.ObjectStorageSecretS3,
		S3: &storage.S3StorageConfig{
			STS:     true,
			Region:  "my-region",
			Buckets: "my-bucket",
		},
		Schemas: []lokiv1.ObjectStorageSchema{
			{
				Version:       lokiv1.ObjectStorageSchemaV11,
				EffectiveDate: "2020-10-01",
			},
		},
	}
	expCfgBytes, err := os.ReadFile("testdata/sts/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)

	opts := defaultOptions()
	opts.ObjectStorage = objStorageConfig

	cfg, _, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
}

func TestBuild_ConfigAndRuntimeConfig_RulerConfigGenerated_WithAlertmanagerClient(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/ruler-with-alertmanager-client/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		Ruler: Ruler{
			Enabled:               true,
			RulesStorageDirectory: "/tmp/rules",
			EvaluationInterval:    "1m",
			PollInterval:          "1m",
			AlertManager: &AlertManagerConfig{
				Notifier: &NotifierConfig{
					TLS: TLSConfig{
						ServerName:         ptr.To("custom-servername"),
						CertPath:           ptr.To("custom/path"),
						KeyPath:            ptr.To("custom/key"),
						CAPath:             ptr.To("custom/CA"),
						InsecureSkipVerify: ptr.To(false),
					},
					BasicAuth: BasicAuth{
						Username: ptr.To("user"),
						Password: ptr.To("pass"),
					},
					HeaderAuth: HeaderAuth{
						CredentialsFile: ptr.To("cred/file"),
						Type:            ptr.To("auth"),
						Credentials:     ptr.To("creds"),
					},
				},
				ExternalURL: "http://alert.me/now",
				ExternalLabels: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
				Hosts:              "http://alerthost1,http://alerthost2",
				EnableV2:           true,
				EnableDiscovery:    true,
				RefreshInterval:    "1m",
				QueueCapacity:      1000,
				Timeout:            "1m",
				ForOutageTolerance: "10m",
				ForGracePeriod:     "5m",
				ResendDelay:        "2m",
			},
			RemoteWrite: &RemoteWriteConfig{
				Enabled:       true,
				RefreshPeriod: "1m",
				Client: &RemoteWriteClientConfig{
					Name:          "remote-write-me",
					URL:           "http://remote.write.me",
					RemoteTimeout: "10s",
					Headers: map[string]string{
						"more": "foryou",
						"less": "forme",
					},
					ProxyURL:        "http://proxy.through.me",
					FollowRedirects: true,
					BearerToken:     "supersecret",
				},
				Queue: &RemoteWriteQueueConfig{
					Capacity:          1000,
					MaxShards:         100,
					MinShards:         50,
					MaxSamplesPerSend: 1000,
					BatchSendDeadline: "10s",
					MinBackOffPeriod:  "30ms",
					MaxBackOffPeriod:  "100ms",
				},
			},
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV11,
					EffectiveDate: "2020-10-01",
				},
			},
		},
		Shippers:              []string{"boltdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}

func TestBuild_ConfigAndRuntimeConfig_OTLPConfigGenerated(t *testing.T) {
	expCfgBytes, err := os.ReadFile("testdata/with-otlp-config/config.yaml")
	require.NoError(t, err)
	expCfg := string(expCfgBytes)
	expRCfg := `
---
overrides:
  test-a:
    otlp_config:
      resource_attributes:
        attributes_config:
        - action: index_label
          attributes:
          - res.foo.bar
          - res.bar.baz
        - action: drop
          attributes:
          - res.service.env
      scope_attributes:
      - action: drop
        attributes:
        - scope.foo.bar
        - scope.bar.baz
      log_attributes:
      - action: drop
        attributes:
        - log.foo.bar
        - log.bar.baz
`
	opts := Options{
		Stack: lokiv1.LokiStackSpec{
			Replication: &lokiv1.ReplicationSpec{
				Factor: 1,
			},
			Limits: &lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					IngestionLimits: &lokiv1.IngestionLimitSpec{
						IngestionRate:             4,
						IngestionBurstSize:        6,
						MaxLabelNameLength:        1024,
						MaxLabelValueLength:       2048,
						MaxLabelNamesPerSeries:    30,
						MaxGlobalStreamsPerTenant: 0,
						MaxLineSize:               256000,
						PerStreamRateLimit:        5,
						PerStreamRateLimitBurst:   15,
						PerStreamDesiredRate:      3,
					},
					QueryLimits: &lokiv1.QueryLimitSpec{
						MaxEntriesLimitPerQuery: 5000,
						MaxChunksPerQuery:       2000000,
						MaxQuerySeries:          500,
						QueryTimeout:            "1m",
						CardinalityLimit:        100000,
						MaxVolumeSeries:         1000,
					},
				},
			},
		},
		Namespace: "test-ns",
		Name:      "test",
		Compactor: Address{
			FQDN: "loki-compactor-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		FrontendWorker: Address{
			FQDN: "loki-query-frontend-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		GossipRing: GossipRing{
			InstancePort:         9095,
			BindPort:             7946,
			MembersDiscoveryAddr: "loki-gossip-ring-lokistack-dev.default.svc.cluster.local",
		},
		Querier: Address{
			Protocol: "http",
			FQDN:     "loki-querier-http-lokistack-dev.default.svc.cluster.local",
			Port:     3100,
		},
		IndexGateway: Address{
			FQDN: "loki-index-gateway-grpc-lokistack-dev.default.svc.cluster.local",
			Port: 9095,
		},
		Ruler: Ruler{
			Enabled:               true,
			RulesStorageDirectory: "/tmp/rules",
			EvaluationInterval:    "1m",
			PollInterval:          "1m",
			AlertManager: &AlertManagerConfig{
				Notifier: &NotifierConfig{
					TLS: TLSConfig{
						ServerName:         ptr.To("custom-servername"),
						CertPath:           ptr.To("custom/path"),
						KeyPath:            ptr.To("custom/key"),
						CAPath:             ptr.To("custom/CA"),
						InsecureSkipVerify: ptr.To(false),
					},
					BasicAuth: BasicAuth{
						Username: ptr.To("user"),
						Password: ptr.To("pass"),
					},
					HeaderAuth: HeaderAuth{
						CredentialsFile: ptr.To("cred/file"),
						Type:            ptr.To("auth"),
						Credentials:     ptr.To("creds"),
					},
				},
				ExternalURL: "http://alert.me/now",
				ExternalLabels: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
				Hosts:              "http://alerthost1,http://alerthost2",
				EnableV2:           true,
				EnableDiscovery:    true,
				RefreshInterval:    "1m",
				QueueCapacity:      1000,
				Timeout:            "1m",
				ForOutageTolerance: "10m",
				ForGracePeriod:     "5m",
				ResendDelay:        "2m",
			},
			RemoteWrite: &RemoteWriteConfig{
				Enabled:       true,
				RefreshPeriod: "1m",
				Client: &RemoteWriteClientConfig{
					Name:          "remote-write-me",
					URL:           "http://remote.write.me",
					RemoteTimeout: "10s",
					Headers: map[string]string{
						"more": "foryou",
						"less": "forme",
					},
					ProxyURL:        "http://proxy.through.me",
					FollowRedirects: true,
					BearerToken:     "supersecret",
				},
				Queue: &RemoteWriteQueueConfig{
					Capacity:          1000,
					MaxShards:         100,
					MinShards:         50,
					MaxSamplesPerSend: 1000,
					BatchSendDeadline: "10s",
					MinBackOffPeriod:  "30ms",
					MaxBackOffPeriod:  "100ms",
				},
			},
		},
		StorageDirectory: "/tmp/loki",
		MaxConcurrent: MaxConcurrent{
			AvailableQuerierCPUCores: 2,
		},
		WriteAheadLog: WriteAheadLog{
			Directory:             "/tmp/wal",
			IngesterMemoryRequest: 4 * 1024 * 1024 * 1024,
		},
		ObjectStorage: storage.Options{
			SharedStore: lokiv1.ObjectStorageSecretS3,
			S3: &storage.S3StorageConfig{
				Endpoint:       "http://test.default.svc.cluster.local.:9000",
				Region:         "us-east",
				Buckets:        "loki",
				ForcePathStyle: true,
			},
			Schemas: []lokiv1.ObjectStorageSchema{
				{
					Version:       lokiv1.ObjectStorageSchemaV13,
					EffectiveDate: "2024-01-01",
				},
			},
			AllowStructuredMetadata: true,
		},
		Shippers:              []string{"tsdb"},
		EnableRemoteReporting: true,
		HTTPTimeouts: HTTPTimeoutConfig{
			IdleTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Minute,
		},
		Overrides: map[string]LokiOverrides{
			"test-a": {
				Limits: lokiv1.PerTenantLimitsTemplateSpec{
					OTLP: &lokiv1.OTLPSpec{
						// This part of the spec is not actually used in this step.
						// It has already been pre-processed into the OTLPAttributes below.
					},
				},
			},
		},
		OTLPAttributes: OTLPAttributeConfig{
			RemoveDefaultLabels: true,
			Global: &OTLPTenantAttributeConfig{
				ResourceAttributes: []OTLPAttribute{
					{
						Action: OTLPAttributeActionStreamLabel,
						Names: []string{
							"res.foo.bar",
							"res.bar.baz",
						},
					},
					{
						Action: OTLPAttributeActionDrop,
						Names: []string{
							"res.service.env",
						},
					},
				},
				ScopeAttributes: []OTLPAttribute{
					{
						Action: OTLPAttributeActionDrop,
						Names: []string{
							"scope.foo.bar",
							"scope.bar.baz",
						},
					},
				},
				LogAttributes: []OTLPAttribute{
					{
						Action: OTLPAttributeActionDrop,
						Names: []string{
							"log.foo.bar",
							"log.bar.baz",
						},
					},
				},
			},
			Tenants: map[string]*OTLPTenantAttributeConfig{
				"test-a": {
					ResourceAttributes: []OTLPAttribute{
						{
							Action: OTLPAttributeActionStreamLabel,
							Names: []string{
								"res.foo.bar",
								"res.bar.baz",
							},
						},
						{
							Action: OTLPAttributeActionDrop,
							Names: []string{
								"res.service.env",
							},
						},
					},
					ScopeAttributes: []OTLPAttribute{
						{
							Action: OTLPAttributeActionDrop,
							Names: []string{
								"scope.foo.bar",
								"scope.bar.baz",
							},
						},
					},
					LogAttributes: []OTLPAttribute{
						{
							Action: OTLPAttributeActionDrop,
							Names: []string{
								"log.foo.bar",
								"log.bar.baz",
							},
						},
					},
				},
			},
		},
	}
	cfg, rCfg, err := Build(opts)
	require.NoError(t, err)
	require.YAMLEq(t, expCfg, string(cfg))
	require.YAMLEq(t, expRCfg, string(rCfg))
}
