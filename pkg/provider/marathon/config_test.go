package marathon

import (
	"context"
	"math"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/types"
	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigurationAPIErrors(t *testing.T) {
	fakeClient := newFakeClient(true, marathon.Applications{})

	p := &Provider{
		marathonClient: fakeClient,
	}

	actualConfig := p.getConfigurations(context.Background())
	fakeClient.AssertExpectations(t)

	if actualConfig != nil {
		t.Errorf("configuration should have been nil, got %v", actualConfig)
	}
}

func TestBuildConfiguration(t *testing.T) {
	testCases := []struct {
		desc                      string
		applications              *marathon.Applications
		constraints               []*types.Constraint
		filterMarathonConstraints bool
		defaultRule               string
		expected                  *config.Configuration
	}{
		{
			desc: "simple application",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80),
					withTasks(localhostTask(taskPorts(80))),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "app",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"app": {LoadBalancer: &config.LoadBalancerService{
							Servers: []config.Server{
								{
									URL: "http://localhost:80",
								},
							},
							PassHostHeader: true,
						}},
					},
				},
			},
		},
		{
			desc: "filtered task",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80),
					withTasks(localhostTask(taskPorts(80), taskState(taskStateStaging))),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "multiple ports",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "app",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"app": {LoadBalancer: &config.LoadBalancerService{
							Servers: []config.Server{
								{
									URL: "http://localhost:80",
								},
							},
							PassHostHeader: true,
						}},
					},
				},
			},
		},
		{
			desc: "with basic auth",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80),
					withLabel("traefik.http.middlewares.Middleware1.basicauth.users", "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
					withLabel("traefik.http.routers.app.middlewares", "Middleware1"),
					withTasks(localhostTask(taskPorts(80))),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service:     "app",
							Rule:        "Host(`app.marathon.localhost`)",
							Middlewares: []string{"Middleware1"},
						},
					},
					Middlewares: map[string]*config.Middleware{
						"Middleware1": {
							BasicAuth: &config.BasicAuth{
								Users: []string{
									"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
									"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
								},
							},
						},
					},
					Services: map[string]*config.Service{
						"app": {LoadBalancer: &config.LoadBalancerService{
							Servers: []config.Server{
								{
									URL: "http://localhost:80",
								},
							},
							PassHostHeader: true,
						}},
					},
				},
			},
		},
		{
			desc: "2 applications in the same service",
			applications: withApplications(
				application(
					appID("/foo-v000"),
					withTasks(localhostTask(taskPorts(8080))),

					withLabel("traefik.http.services.Service1.LoadBalancer.server.port", "index:0"),
					withLabel("traefik.http.routers.Router1.rule", "Host(`app.marathon.localhost`)"),
				),
				application(
					appID("/foo-v001"),
					withTasks(localhostTask(taskPorts(8081))),

					withLabel("traefik.http.services.Service1.LoadBalancer.server.port", "index:0"),
					withLabel("traefik.http.routers.Router1.rule", "Host(`app.marathon.localhost`)"),
				),
			),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Router1": {
							Service: "Service1",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {LoadBalancer: &config.LoadBalancerService{
							Servers: []config.Server{
								{
									URL: "http://localhost:8080",
								},
								{
									URL: "http://localhost:8081",
								},
							},
							PassHostHeader: true,
						}},
					},
				},
			},
		},
		{
			desc: "2 applications with 2 tasks in the same service",
			applications: withApplications(
				application(
					appID("/foo-v000"),
					withTasks(localhostTask(taskPorts(8080))),
					withTasks(localhostTask(taskPorts(8081))),

					withLabel("traefik.http.services.Service1.LoadBalancer.server.port", "index:0"),
					withLabel("traefik.http.routers.Router1.rule", "Host(`app.marathon.localhost`)"),
				),
				application(
					appID("/foo-v001"),
					withTasks(localhostTask(taskPorts(8082))),
					withTasks(localhostTask(taskPorts(8083))),

					withLabel("traefik.http.services.Service1.LoadBalancer.server.port", "index:0"),
					withLabel("traefik.http.routers.Router1.rule", "Host(`app.marathon.localhost`)"),
				),
			),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Router1": {
							Service: "Service1",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {LoadBalancer: &config.LoadBalancerService{
							Servers: []config.Server{
								{
									URL: "http://localhost:8080",
								},
								{
									URL: "http://localhost:8081",
								},
								{
									URL: "http://localhost:8082",
								},
								{
									URL: "http://localhost:8083",
								},
							},
							PassHostHeader: true,
						}},
					},
				},
			},
		},
		{
			desc: "2 applications",
			applications: withApplications(
				application(
					appID("/foo"),
					withTasks(localhostTask(taskPorts(8080))),
				),
				application(
					appID("/bar"),
					withTasks(localhostTask(taskPorts(8081))),
				),
			),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"foo": {
							Service: "foo",
							Rule:    "Host(`foo.marathon.localhost`)",
						},
						"bar": {
							Service: "bar",
							Rule:    "Host(`bar.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"foo": {LoadBalancer: &config.LoadBalancerService{
							Servers: []config.Server{
								{
									URL: "http://localhost:8080",
								},
							},
							PassHostHeader: true,
						}},
						"bar": {LoadBalancer: &config.LoadBalancerService{
							Servers: []config.Server{
								{
									URL: "http://localhost:8081",
								},
							},
							PassHostHeader: true,
						}},
					},
				},
			},
		},
		{
			desc: "two tasks no labels",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80),
					withTasks(localhostTask(taskPorts(80)), localhostTask(taskPorts(81))),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "app",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"app": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
									{
										URL: "http://localhost:81",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "simple application with label on service",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80),
					withTasks(localhostTask(taskPorts(80))),
					withLabel("traefik.http.services.Service1.loadbalancer.passhostheader", "true"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "Service1",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {LoadBalancer: &config.LoadBalancerService{
							Servers: []config.Server{
								{
									URL: "http://localhost:80",
								},
							},
							PassHostHeader: true,
						}},
					},
				},
			},
		},
		{
			desc: "one app with labels",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.services.Service1.loadbalancer.passhostheader", "true"),
					withLabel("traefik.http.routers.Router1.rule", "Host(`foo.com`)"),
					withLabel("traefik.http.routers.Router1.service", "Service1"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Router1": {
							Service: "Service1",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one app with rule label",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.routers.Router1.rule", "Host(`foo.com`)"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"app": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
					Routers: map[string]*config.Router{
						"Router1": {
							Service: "app",
							Rule:    "Host(`foo.com`)",
						},
					},
				},
			},
		},
		{
			desc: "one app with rule label and one service",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.routers.Router1.rule", "Host(`foo.com`)"),
					withLabel("traefik.http.services.Service1.loadbalancer.passhostheader", "true"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Router1": {
							Service: "Service1",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one app with rule label and two services",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.routers.Router1.rule", "Host(`foo.com`)"),
					withLabel("traefik.http.services.Service1.loadbalancer.passhostheader", "true"),
					withLabel("traefik.http.services.Service2.loadbalancer.passhostheader", "true"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"Service2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "two apps with same service name and different passhostheader",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.services.Service1.loadbalancer.passhostheader", "false"),
				),
				application(
					appID("/app2"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.services.Service1.loadbalancer.passhostheader", "true"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "Service1",
							Rule:    "Host(`app.marathon.localhost`)",
						},
						"app2": {
							Service: "Service1",
							Rule:    "Host(`app2.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "two apps with two identical middleware",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.middlewares.Middleware1.maxconn.amount", "42"),
				),
				application(
					appID("/app2"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.middlewares.Middleware1.maxconn.amount", "42"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "app",
							Rule:    "Host(`app.marathon.localhost`)",
						},
						"app2": {
							Service: "app2",
							Rule:    "Host(`app2.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{
						"Middleware1": {
							MaxConn: &config.MaxConn{
								Amount:        42,
								ExtractorFunc: "request.host",
							},
						},
					},
					Services: map[string]*config.Service{
						"app": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"app2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "two apps with two different middlewares",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.middlewares.Middleware1.maxconn.amount", "42"),
				),
				application(
					appID("/app2"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.middlewares.Middleware1.maxconn.amount", "41"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "app",
							Rule:    "Host(`app.marathon.localhost`)",
						},
						"app2": {
							Service: "app2",
							Rule:    "Host(`app2.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"app": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"app2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "two apps with two different routers with same name",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.routers.Router1.rule", "Host(`foo.com`)"),
				),
				application(
					appID("/app2"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.routers.Router1.rule", "Host(`bar.com`)"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"app": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"app2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "two apps with two identical routers with same name",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.routers.Router1.rule", "Host(`foo.com`)"),
					withLabel("traefik.http.services.Service1.LoadBalancer.passhostheader", "true"),
				),
				application(
					appID("/app2"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.routers.Router1.rule", "Host(`foo.com`)"),
					withLabel("traefik.http.services.Service1.LoadBalancer.passhostheader", "true"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Router1": {
							Service: "Service1",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "two apps with two identical routers with same name",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.routers.Router1.rule", "Host(`foo.com`)"),
				),
				application(
					appID("/app2"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.routers.Router1.rule", "Host(`foo.com`)"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"app": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"app2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one app with wrong label",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.wrong.label", "tchouk"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "app",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"app": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one app with label port",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.services.Service1.LoadBalancer.server.scheme", "h2c"),
					withLabel("traefik.http.services.Service1.LoadBalancer.server.port", "90"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "Service1",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "h2c://localhost:90",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one app with label port on two services",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.http.services.Service1.LoadBalancer.server.port", ""),
					withLabel("traefik.http.services.Service2.LoadBalancer.server.port", "8080"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"Service2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:8080",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one app without port",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask()),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one app without port with middleware",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask()),
					withLabel("traefik.http.middlewares.Middleware1.basicauth.users", "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one app with traefik.enable=false",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask()),
					withLabel("traefik.enable", "false"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one app with traefik.enable=false",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask()),
					withLabel("traefik.enable", "false"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one app with non matching constraint",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.tags", "foo"),
				)),
			constraints: []*types.Constraint{
				{
					Key:       "tag",
					MustMatch: true,
					Value:     "bar",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one app with non matching marathon constraint",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					constraint("rack_id:CLUSTER:rack-1"),
				)),
			filterMarathonConstraints: true,
			constraints: []*types.Constraint{
				{
					Key:       "tag",
					MustMatch: true,
					Value:     "rack_id:CLUSTER:rack-2",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one app with matching marathon constraint",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					constraint("rack_id:CLUSTER:rack-1"),
				)),
			filterMarathonConstraints: true,
			constraints: []*types.Constraint{
				{
					Key:       "tag",
					MustMatch: true,
					Value:     "rack_id:CLUSTER:rack-1",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "app",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"app": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one app with matching constraint",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.tags", "bar"),
				)),

			constraints: []*types.Constraint{
				{
					Key:       "tag",
					MustMatch: true,
					Value:     "bar",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "app",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"app": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:        "one app with group as subdomain rule",
			defaultRule: `Host("{{ .Name | trimPrefix "/" | splitList "/" | strsToItfs | reverse | join "." }}.marathon.localhost")`,
			applications: withApplications(
				application(
					appID("/a/b/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"a_b_app": {
							Service: "a_b_app",
							Rule:    `Host("app.b.a.marathon.localhost")`,
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"a_b_app": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one app with tcp labels",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.tcp.routers.foo.rule", "HostSNI(`foo.bar`)"),
					withLabel("traefik.tcp.routers.foo.tls", "true"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							Service: "app",
							Rule:    "HostSNI(`foo.bar`)",
							TLS:     &config.RouterTCPTLSConfig{},
						},
					},
					Services: map[string]*config.TCPService{
						"app": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "localhost:80",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one app with tcp labels without rule",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.tcp.routers.foo.tls", "true"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{
						"app": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "localhost:80",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one app with tcp labels with port",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.tcp.routers.foo.rule", "HostSNI(`foo.bar`)"),
					withLabel("traefik.tcp.routers.foo.tls", "true"),
					withLabel("traefik.tcp.services.foo.loadbalancer.server.port", "8080"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							Service: "foo",
							Rule:    "HostSNI(`foo.bar`)",
							TLS:     &config.RouterTCPTLSConfig{},
						},
					},
					Services: map[string]*config.TCPService{
						"foo": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "localhost:8080",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one app with tcp labels with port and http service",
			applications: withApplications(
				application(
					appID("/app"),
					appPorts(80, 81),
					withTasks(localhostTask(taskPorts(80, 81))),
					withLabel("traefik.tcp.routers.foo.rule", "HostSNI(`foo.bar`)"),
					withLabel("traefik.tcp.routers.foo.tls", "true"),
					withLabel("traefik.tcp.services.foo.loadbalancer.server.port", "8080"),
					withLabel("traefik.http.services.bar.loadbalancer.passhostheader", "true"),
				)),
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							Service: "foo",
							Rule:    "HostSNI(`foo.bar`)",
							TLS:     &config.RouterTCPTLSConfig{},
						},
					},
					Services: map[string]*config.TCPService{
						"foo": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "localhost:8080",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"app": {
							Service: "bar",
							Rule:    "Host(`app.marathon.localhost`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"bar": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://localhost:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			defaultRule := "Host(`{{ normalize .Name }}.marathon.localhost`)"
			if len(test.defaultRule) > 0 {
				defaultRule = test.defaultRule
			}

			p := &Provider{
				DefaultRule:               defaultRule,
				ExposedByDefault:          true,
				FilterMarathonConstraints: test.filterMarathonConstraints,
			}
			p.Constraints = test.constraints

			err := p.Init()
			require.NoError(t, err)

			actualConfig := p.buildConfiguration(context.Background(), test.applications)

			assert.NotNil(t, actualConfig)
			assert.Equal(t, test.expected, actualConfig)
		})
	}
}

func TestApplicationFilterEnabled(t *testing.T) {
	testCases := []struct {
		desc             string
		exposedByDefault bool
		enabledLabel     string
		expected         bool
	}{
		{
			desc:             "exposed and tolerated by valid label value",
			exposedByDefault: true,
			enabledLabel:     "true",
			expected:         true,
		},
		{
			desc:             "exposed but overridden by label",
			exposedByDefault: true,
			enabledLabel:     "false",
			expected:         false,
		},
		{
			desc:             "non-exposed but overridden by label",
			exposedByDefault: false,
			enabledLabel:     "true",
			expected:         true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			provider := &Provider{ExposedByDefault: true}

			app := application(withLabel("traefik.enable", test.enabledLabel))

			extraConf, err := provider.getConfiguration(app)
			require.NoError(t, err)

			if provider.keepApplication(context.Background(), extraConf) != test.expected {
				t.Errorf("got unexpected filtering = %t", !test.expected)
			}
		})
	}
}

func TestGetServer(t *testing.T) {
	type expected struct {
		server config.Server
		error  string
	}

	testCases := []struct {
		desc          string
		provider      Provider
		app           marathon.Application
		extraConf     configuration
		defaultServer config.Server
		expected      expected
	}{
		{
			desc:          "undefined host",
			provider:      Provider{},
			app:           application(),
			extraConf:     configuration{},
			defaultServer: config.Server{},
			expected: expected{
				error: `host is undefined for task "taskID" app ""`,
			},
		},
		{
			desc:     "with task port",
			provider: Provider{},
			app: application(
				appID("/app"),
				appPorts(80),
				withTasks(localhostTask(taskPorts(80))),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
			},
			expected: expected{
				server: config.Server{
					URL: "http://localhost:80",
				},
			},
		},
		{
			desc:     "without task port",
			provider: Provider{},
			app: application(
				appID("/app"),
				appPorts(80),
				withTasks(localhostTask()),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
			},
			expected: expected{
				error: "unable to process ports for /app taskID: no port found",
			},
		},
		{
			desc:     "with default server port",
			provider: Provider{},
			app: application(
				appID("/app"),
				appPorts(80),
				withTasks(localhostTask(taskPorts(80))),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
				Port:   "88",
			},
			expected: expected{
				server: config.Server{
					URL: "http://localhost:88",
				},
			},
		},
		{
			desc:     "with invalid default server port",
			provider: Provider{},
			app: application(
				appID("/app"),
				appPorts(80),
				withTasks(localhostTask(taskPorts(80))),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
				Port:   "aaaa",
			},
			expected: expected{
				error: `unable to process ports for /app taskID: strconv.Atoi: parsing "aaaa": invalid syntax`,
			},
		},
		{
			desc:     "with negative default server port",
			provider: Provider{},
			app: application(
				appID("/app"),
				appPorts(80),
				withTasks(localhostTask(taskPorts(80))),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
				Port:   "-6",
			},
			expected: expected{
				error: `unable to process ports for /app taskID: explicitly specified port -6 must be greater than zero`,
			},
		},
		{
			desc:     "with port index",
			provider: Provider{},
			app: application(
				appID("/app"),
				appPorts(80),
				withTasks(localhostTask(taskPorts(80, 81))),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
				Port:   "index:1",
			},
			expected: expected{
				server: config.Server{
					URL: "http://localhost:81",
				},
			},
		},
		{
			desc:     "with out of range port index",
			provider: Provider{},
			app: application(
				appID("/app"),
				appPorts(80),
				withTasks(localhostTask(taskPorts(80, 81))),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
				Port:   "index:2",
			},
			expected: expected{
				error: "unable to process ports for /app taskID: index 2 must be within range (0, 1)",
			},
		},
		{
			desc:     "with invalid port index",
			provider: Provider{},
			app: application(
				appID("/app"),
				appPorts(80),
				withTasks(localhostTask(taskPorts(80, 81))),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
				Port:   "index:aaa",
			},
			expected: expected{
				error: `unable to process ports for /app taskID: strconv.Atoi: parsing "aaa": invalid syntax`,
			},
		},
		{
			desc:     "with application port and no task port",
			provider: Provider{},
			app: application(
				appID("/app"),
				appPorts(80),
				portDefinition(80),
				withTasks(localhostTask()),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
			},
			expected: expected{
				server: config.Server{
					URL: "http://localhost:80",
				},
			},
		},
		{
			desc:     "with IP per task",
			provider: Provider{},
			app: application(
				appID("/app"),
				appPorts(80),
				ipAddrPerTask(88),
				withTasks(localhostTask()),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
			},
			expected: expected{
				server: config.Server{
					URL: "http://127.0.0.1:88",
				},
			},
		},
		{
			desc:     "with container network",
			provider: Provider{},
			app: application(
				containerNetwork(),
				appID("/app"),
				appPorts(80),
				withTasks(localhostTask(taskPorts(80, 81))),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
			},
			expected: expected{
				server: config.Server{
					URL: "http://127.0.0.1:80",
				},
			},
		},
		{
			desc:     "with bridge network",
			provider: Provider{},
			app: application(
				bridgeNetwork(),
				appID("/app"),
				appPorts(83),
				withTasks(localhostTask(taskPorts(80, 81))),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
			},
			expected: expected{
				server: config.Server{
					URL: "http://localhost:80",
				},
			},
		},
		{
			desc:     "with several IP addresses on task",
			provider: Provider{},
			app: application(
				ipAddrPerTask(88),
				appID("/app"),
				appPorts(83),
				withTasks(
					task(
						withTaskID("myTask"),
						host("localhost"),
						ipAddresses("127.0.0.1", "127.0.0.2"),
						taskState(taskStateRunning),
					)),
			),
			extraConf: configuration{
				Marathon: specificConfiguration{
					IPAddressIdx: 0,
				},
			},
			defaultServer: config.Server{
				Scheme: "http",
			},
			expected: expected{
				server: config.Server{
					URL: "http://127.0.0.1:88",
				},
			},
		},
		{
			desc:     "with several IP addresses on task, undefined [MinInt32] IPAddressIdx",
			provider: Provider{},
			app: application(
				ipAddrPerTask(88),
				appID("/app"),
				appPorts(83),
				withTasks(
					task(
						host("localhost"),
						ipAddresses("127.0.0.1", "127.0.0.2"),
						taskState(taskStateRunning),
					)),
			),
			extraConf: configuration{
				Marathon: specificConfiguration{
					IPAddressIdx: math.MinInt32,
				},
			},
			defaultServer: config.Server{
				Scheme: "http",
			},
			expected: expected{
				error: "found 2 task IP addresses but missing IP address index for Marathon application /app on task taskID",
			},
		},
		{
			desc:     "with several IP addresses on task, IPAddressIdx out of range",
			provider: Provider{},
			app: application(
				ipAddrPerTask(88),
				appID("/app"),
				appPorts(83),
				withTasks(
					task(
						host("localhost"),
						ipAddresses("127.0.0.1", "127.0.0.2"),
						taskState(taskStateRunning),
					)),
			),
			extraConf: configuration{
				Marathon: specificConfiguration{
					IPAddressIdx: 3,
				},
			},
			defaultServer: config.Server{
				Scheme: "http",
			},
			expected: expected{
				error: "cannot use IP address index to select from 2 task IP addresses for Marathon application /app on task taskID",
			},
		},
		{
			desc:     "with task without IP address",
			provider: Provider{},
			app: application(
				ipAddrPerTask(88),
				appID("/app"),
				appPorts(83),
				withTasks(
					task(
						host("localhost"),
						taskState(taskStateRunning),
					)),
			),
			extraConf: configuration{},
			defaultServer: config.Server{
				Scheme: "http",
			},
			expected: expected{
				error: "missing IP address for Marathon application /app on task taskID",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			task := task()
			if len(test.app.Tasks) > 0 {
				task = *test.app.Tasks[0]
			}

			server, err := test.provider.getServer(test.app, task, test.extraConf, test.defaultServer)
			if len(test.expected.error) > 0 {
				require.EqualError(t, err, test.expected.error)
			} else {
				require.NoError(t, err)

				assert.Equal(t, test.expected.server, server)
			}
		})
	}
}
