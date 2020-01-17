/*
/*
@Time : 2020/1/7 3:55 下午
@Author : tianpeng.du
@File : endpoint.go
@Software: GoLand
*/
package main

import (
	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
)

func NewEdsClusterConfig(clusterName string, adsClusterName string) *api.Cluster_EdsClusterConfig {
	return &api.Cluster_EdsClusterConfig{
		EdsConfig: &v2_core.ConfigSource{
			ConfigSourceSpecifier: &v2_core.ConfigSource_ApiConfigSource{
				ApiConfigSource: &v2_core.ApiConfigSource{
					ApiType: v2_core.ApiConfigSource_GRPC,
					GrpcServices: []*v2_core.GrpcService{
						&v2_core.GrpcService{
							TargetSpecifier: &v2_core.GrpcService_EnvoyGrpc_{
								EnvoyGrpc: &v2_core.GrpcService_EnvoyGrpc{
									ClusterName: "ads_cluster",
								},
							},
						},
					},
				},
			},
		},
		ServiceName: "local_cluster",
	}
}

func NewLoadAssignment(clusterName string, addrs []Addr) *api.ClusterLoadAssignment {
	lbEndpoints := make([]*endpoint.LbEndpoint, len(addrs))
	for k, addr := range addrs {
		lbEndpoints[k] = &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &v2_core.Address{
						Address: &v2_core.Address_SocketAddress{
							SocketAddress: &v2_core.SocketAddress{
								Protocol: v2_core.SocketAddress_TCP,
								Address:  addr.Address,
								PortSpecifier: &v2_core.SocketAddress_PortValue{
									PortValue: addr.Port,
								},
							},
						},
					},
				},
			},
		}
	}
	return &api.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{
			&endpoint.LocalityLbEndpoints{
				LbEndpoints: lbEndpoints,
			},
		},
	}
}
