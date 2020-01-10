/*
/*
@Time : 2020/1/7 3:46 下午
@Author : tianpeng.du
@File : cluster.go
@Software: GoLand
*/
package main

import (
	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/ptypes/duration"
)

func NewEdsCluster(clusterName string,eds *api.Cluster_EdsClusterConfig) cache.Resource {
	return &api.Cluster{
		Name:           clusterName,
		AltStatName:    clusterName,
		ConnectTimeout: &duration.Duration{Seconds: 1},
		ClusterDiscoveryType: &api.Cluster_Type{
			Type: api.Cluster_EDS,
		},
		LbPolicy: api.Cluster_ROUND_ROBIN,
		EdsClusterConfig: eds,
	}
}

func NewStaticCluster(name string,endpoints *api.ClusterLoadAssignment) cache.Resource {
	return &api.Cluster{
		Name:        name,
		AltStatName: name,
		ClusterDiscoveryType: &api.Cluster_Type{
			Type: api.Cluster_STATIC,
		},
		EdsClusterConfig:              nil,
		ConnectTimeout:                &duration.Duration{Seconds: 1},
		PerConnectionBufferLimitBytes: nil, //default 1MB
		LbPolicy:                      api.Cluster_ROUND_ROBIN,
		LoadAssignment:                endpoints,
	}
}