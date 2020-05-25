/*
/*
@Time : 2020/1/3 3:02 下午
@Author : tianpeng.du
@File : main.go
@Software: GoLand
*/
package main

import (
	"context"
	"fmt"
	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	frl "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/rate_limit/v2"
	http_router "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/router/v2"
	http_conn_manager "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	rl "github.com/envoyproxy/go-control-plane/envoy/config/ratelimit/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/conversion"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	snapshotCache := cache.NewSnapshotCache(false, cache.IDHash{}, nil)
	server := xds.NewServer(ctx, snapshotCache, nil)
	grpcServer := grpc.NewServer()
	lis, _ := net.Listen("tcp", ":8080")

	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	api.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	api.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	api.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	api.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	node := v2_core.Node{
		Id:                   "1",
		Cluster:              "envoy-test",
		Metadata:             nil,
		Locality:             nil,
		BuildVersion:         "",
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}
	endpoints := []cache.Resource{}
	clusters := []cache.Resource{}
	routes := []cache.Resource{}
	listeners := []cache.Resource{}

	clusters = append(clusters, &api.Cluster{
		Name:           "local_cluster",
		AltStatName:    "local_cluster",
		ConnectTimeout: &duration.Duration{Seconds: 1},
		ClusterDiscoveryType: &api.Cluster_Type{
			Type: api.Cluster_EDS,
		},
		LbPolicy: api.Cluster_ROUND_ROBIN,
		EdsClusterConfig: &api.Cluster_EdsClusterConfig{
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
				}, // 使用EDS
				//ConfigSourceSpecifier: &v2_core.ConfigSource_Ads{
				//	Ads: &v2_core.AggregatedConfigSource{}, //使用ADS
				//},
			},
			ServiceName: "local_cluster",
		},
	})

	clusters = append(clusters, &api.Cluster{
		Name:           "rate_limit_cluster",
		AltStatName:    "rate_limit_cluster",
		ConnectTimeout: &duration.Duration{Seconds: 1},
		ClusterDiscoveryType: &api.Cluster_Type{
			Type: api.Cluster_EDS,
		},
		LbPolicy: api.Cluster_ROUND_ROBIN,
		Http2ProtocolOptions: &v2_core.Http2ProtocolOptions{},
		EdsClusterConfig: &api.Cluster_EdsClusterConfig{
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
				}, // 使用EDS
				//ConfigSourceSpecifier: &v2_core.ConfigSource_Ads{
				//	Ads: &v2_core.AggregatedConfigSource{}, //使用ADS
				//},
			},
			ServiceName: "rate_limit_cluster",
		},
	})

	endpoints = append(endpoints, &api.ClusterLoadAssignment{
		ClusterName: "local_cluster",
		Endpoints: []*endpoint.LocalityLbEndpoints{
			&endpoint.LocalityLbEndpoints{
				LbEndpoints: []*endpoint.LbEndpoint{
					&endpoint.LbEndpoint{
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: &v2_core.Address{
									Address: &v2_core.Address_SocketAddress{
										SocketAddress: &v2_core.SocketAddress{
											Protocol: v2_core.SocketAddress_TCP,
											Address:  "127.0.0.1",
											PortSpecifier: &v2_core.SocketAddress_PortValue{
												PortValue: 8000,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	endpoints = append(endpoints, &api.ClusterLoadAssignment{
		ClusterName: "rate_limit_cluster",
		Endpoints: []*endpoint.LocalityLbEndpoints{
			&endpoint.LocalityLbEndpoints{
				LbEndpoints: []*endpoint.LbEndpoint{
					&endpoint.LbEndpoint{
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: &v2_core.Address{
									Address: &v2_core.Address_SocketAddress{
										SocketAddress: &v2_core.SocketAddress{
											Protocol: v2_core.SocketAddress_TCP,
											Address:  "127.0.0.1",
											PortSpecifier: &v2_core.SocketAddress_PortValue{
												PortValue: 8081,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	http_filter_router_ := &http_router.Router{
		DynamicStats: &wrappers.BoolValue{
			Value: true,
		},
	}
	http_filter_router, err := conversion.MessageToStruct(http_filter_router_)

	frlo_ := &frl.RateLimit{
		Domain:                         "fuckcontour",
		Stage:                          0,
		RequestType:                    "external",
		FailureModeDeny:                false,
		RateLimitedAsResourceExhausted: false,
		RateLimitService: &rl.RateLimitServiceConfig{
			GrpcService: &v2_core.GrpcService{
				TargetSpecifier: &v2_core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &v2_core.GrpcService_EnvoyGrpc{
						ClusterName: "rate_limit_cluster",
					},
				},
			},
		},
	}
	frlo, err := ptypes.MarshalAny(frlo_)
	listen_filter_http_conn_ := &http_conn_manager.HttpConnectionManager{
		StatPrefix: "ingress_http",
		RouteSpecifier: &http_conn_manager.HttpConnectionManager_Rds{
			Rds: &http_conn_manager.Rds{
				RouteConfigName: "test-route",
				ConfigSource: &v2_core.ConfigSource{
					//ConfigSourceSpecifier: &v2_core.ConfigSource_ApiConfigSource{
					//	ApiConfigSource: &v2_core.ApiConfigSource{
					//		ApiType: v2_core.ApiConfigSource_GRPC,
					//		GrpcServices: []*v2_core.GrpcService{
					//			&v2_core.GrpcService{
					//				TargetSpecifier: &v2_core.GrpcService_EnvoyGrpc_{
					//					EnvoyGrpc: &v2_core.GrpcService_EnvoyGrpc{
					//						ClusterName: "ads_cluster",
					//					},
					//				},
					//			},
					//		},
					//	},
					//}, 使用RDS
					ConfigSourceSpecifier: &v2_core.ConfigSource_Ads{
						Ads: &v2_core.AggregatedConfigSource{}, //使用ADS
					},
				},
			},
		},
		HttpFilters: []*http_conn_manager.HttpFilter{
			&http_conn_manager.HttpFilter{
				Name:       "envoy.rate_limit",
				ConfigType: &http_conn_manager.HttpFilter_TypedConfig{TypedConfig: frlo},
			},
			&http_conn_manager.HttpFilter{
				Name:       "envoy.router",
				ConfigType: &http_conn_manager.HttpFilter_Config{Config: http_filter_router},
			},
		},
	}

	listen_filter_http_conn, err := conversion.MessageToStruct(listen_filter_http_conn_)
	if err != nil {
		log.Println(err)
	}

	routes = append(routes, &api.RouteConfiguration{
		Name: "test-route",
		VirtualHosts: []*route.VirtualHost{
			&route.VirtualHost{
				Name: "local",
				Domains: []string{
					"*",
				},
				Routes: []*route.Route{
					&route.Route{
						Match: &route.RouteMatch{
							PathSpecifier: &route.RouteMatch_Prefix{
								Prefix: "/",
							},
							CaseSensitive: &wrappers.BoolValue{
								Value: false,
							},
						},
						Action: &route.Route_Route{
							Route: &route.RouteAction{
								ClusterSpecifier: &route.RouteAction_Cluster{
									Cluster: "local_cluster",
								},
								HostRewriteSpecifier: &route.RouteAction_HostRewrite{
									HostRewrite: "127.0.0.1",
								},
							},
						},
					},
				},
				//---
				RateLimits: []*route.RateLimit{
					&route.RateLimit{
						Stage: &wrappers.UInt32Value{Value: 0},
						Actions: []*route.RateLimit_Action{
							&route.RateLimit_Action{
								ActionSpecifier: &route.RateLimit_Action_GenericKey_{
									GenericKey: &route.RateLimit_Action_GenericKey{
										DescriptorValue: "apis",
									},
								},
							},
						},
					},
				},
				//---
			},
		},
	})
	listeners = append(listeners, &api.Listener{
		Name: "test",
		Address: &v2_core.Address{
			Address: &v2_core.Address_SocketAddress{
				SocketAddress: &v2_core.SocketAddress{
					Protocol: v2_core.SocketAddress_TCP,
					Address:  "127.0.0.1",
					PortSpecifier: &v2_core.SocketAddress_PortValue{
						PortValue: 7777,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{
			&listener.FilterChain{
				Filters: []*listener.Filter{
					&listener.Filter{
						Name: "envoy.http_connection_manager",
						ConfigType: &listener.Filter_Config{
							Config: listen_filter_http_conn,
						},
					},
				},
			},
		},
	})
	err = snapshotCache.SetSnapshot(node.Id, cache.NewSnapshot("1", endpoints, clusters, routes, listeners, nil))
	if err != nil {
		log.Println(err)
	}
	input := ""
	fmt.Scanf("\n", &input)
	log.Println("new version")
	routes = []cache.Resource{&api.RouteConfiguration{
		Name: "test-route",
		VirtualHosts: []*route.VirtualHost{
			&route.VirtualHost{
				Name: "local",
				Domains: []string{
					"*",
				},
				RateLimits: []*route.RateLimit{
					&route.RateLimit{
						Stage: &wrappers.UInt32Value{Value: 0},
						Actions: []*route.RateLimit_Action{
							&route.RateLimit_Action{
								ActionSpecifier: &route.RateLimit_Action_GenericKey_{
									GenericKey: &route.RateLimit_Action_GenericKey{
										DescriptorValue: "apis",
									},
								},
							},
						},
					},
				},
				Routes: []*route.Route{
					&route.Route{
						Match: &route.RouteMatch{
							PathSpecifier: &route.RouteMatch_Prefix{
								Prefix: "/",
							},
							CaseSensitive: &wrappers.BoolValue{
								Value: false,
							},
						},
						Action: &route.Route_Route{
							Route: &route.RouteAction{
								ClusterSpecifier: &route.RouteAction_Cluster{
									Cluster: "local_cluster",
								},
								HostRewriteSpecifier: &route.RouteAction_HostRewrite{
									HostRewrite: "127.0.0.1",
								},
							},
						},
					},
				},
			},
		},
	}}

	err = snapshotCache.SetSnapshot(node.Id, cache.NewSnapshot("2", endpoints, clusters, routes, listeners, nil))
	if err != nil {
		log.Println(err)
	}
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Println(err)
		}
	}()
	log.Println(123)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	select {
	case <-c:
		cancel()
	}
}
