package docker

import (
	"context"
	"strconv"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRule(t *testing.T) {
	testCases := []struct {
		desc        string
		containers  []dockerData
		defaultRule string
		expected    *config.Configuration
	}{
		{
			desc: "default rule with no variable",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: "Host(`foo.bar`)",
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`foo.bar`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "default rule with service name",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: "Host(`{{ .Name }}.foo.bar`)",
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.foo.bar`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "default rule with label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.domain": "foo.bar",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: `Host("{{ .Name }}.{{ index .Labels "traefik.domain" }}")`,
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    `Host("Test.foo.bar")`,
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "invalid rule",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: `Host("{{ .Toto }}")`,
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "undefined rule",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: ``,
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "default template rule",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: DefaultTemplateRule,
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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

			p := Provider{
				ExposedByDefault: true,
				DefaultRule:      test.defaultRule,
			}

			err := p.Init()
			require.NoError(t, err)

			for i := 0; i < len(test.containers); i++ {
				var err error
				test.containers[i].ExtraConf, err = p.getConfiguration(test.containers[i])
				require.NoError(t, err)
			}

			configuration := p.buildConfiguration(context.Background(), test.containers)

			assert.Equal(t, test.expected, configuration)
		})
	}
}

func Test_buildConfiguration(t *testing.T) {
	testCases := []struct {
		desc        string
		containers  []dockerData
		constraints []*types.Constraint
		expected    *config.Configuration
	}{
		{
			desc: "one container no label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "two containers no label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ServiceName: "Test2",
					Name:        "Test2",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
						"Test2": {
							Service: "Test2",
							Rule:    "Host(`Test2.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"Test2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.2:80",
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
			desc: "two containers with same service name no label",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
			desc: "one container with label (not on server)",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "one container with labels",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.routers.Router1.service":                       "Service1",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
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
										URL: "http://127.0.0.1:80",
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
			desc: "one container with rule label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
					Routers: map[string]*config.Router{
						"Router1": {
							Service: "Test",
							Rule:    "Host(`foo.com`)",
						},
					},
				},
			},
		},
		{
			desc: "one container with rule label and one service",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
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
										URL: "http://127.0.0.1:80",
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
			desc: "one container with rule label and two services",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
						"traefik.http.services.Service2.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"Service2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "two containers with same service name and different passhostheader",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "false",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "three containers with same service name and different passhostheader",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "false",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
				{
					ID:          "3",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "two containers with same service name and same LB methods",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
			desc: "one container with MaxConn in label (default value)",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.maxconn.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
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
				},
			},
		},
		{
			desc: "two containers with two identical middlewares",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.maxconn.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.maxconn.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
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
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
			desc: "two containers with two different middlewares with same name",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.maxconn.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.maxconn.amount": "41",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
			desc: "three containers with different middlewares with same name",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.maxconn.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.maxconn.amount": "41",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
				{
					ID:          "3",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.maxconn.amount": "40",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.3",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
									{
										URL: "http://127.0.0.3:80",
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
			desc: "two containers with two different routers with same name",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`bar.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
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
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
			desc: "three containers with different routers with same name",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`bar.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
				{
					ID:          "3",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foobar.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.3",
							},
						},
					},
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
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
									{
										URL: "http://127.0.0.3:80",
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
			desc: "two containers with two identical routers",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Router1": {
							Service: "Test",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
			desc: "two containers with two identical router rules and different service names",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ServiceName: "Test2",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
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
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"Test2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.2:80",
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
			desc: "one container with bad label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.wrong.label": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "one container with label port",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.scheme": "h2c",
						"traefik.http.services.Service1.LoadBalancer.server.port":   "8080",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "h2c://127.0.0.1:8080",
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
			desc: "one container with label port on two services",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.port": "",
						"traefik.http.services.Service2.LoadBalancer.server.port": "8080",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"Service2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:8080",
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
			desc: "one container without port",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			desc: "one container without port with middleware",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.maxconn.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			desc: "one container with traefik.enable false",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.enable": "false",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			desc: "one container not healthy",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
					Health: "not_healthy",
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
			desc: "one container with non matching constraints",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tags": "foo",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
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
			desc: "one container with matching constraints",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tags": "foo",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			constraints: []*types.Constraint{
				{
					Key:       "tag",
					MustMatch: true,
					Value:     "foo",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "Middlewares used in router",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.basicauth.users": "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						"traefik.http.routers.Test.middlewares":                "Middleware1",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
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
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
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
			desc: "tcp with label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule": "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":  "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							Service: "Test",
							Rule:    "HostSNI(`foo.bar`)",
							TLS:     &config.RouterTCPTLSConfig{},
						},
					},
					Services: map[string]*config.TCPService{
						"Test": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:80",
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
			desc: "tcp with label without rule",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.tls": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{
						"Test": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:80",
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
			desc: "tcp with label and port",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule":                      "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":                       "true",
						"traefik.tcp.services.foo.loadbalancer.server.port": "8080",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
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
										Address: "127.0.0.1:8080",
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
			desc: "tcp with label and port and http service",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule":                               "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":                                "true",
						"traefik.tcp.services.foo.loadbalancer.server.port":          "8080",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule":                               "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":                                "true",
						"traefik.tcp.services.foo.loadbalancer.server.port":          "8080",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
			},
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
										Address: "127.0.0.1:8080",
									},
									{
										Address: "127.0.0.2:8080",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
			desc: "tcp with label for tcp service",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.services.foo.loadbalancer.server.port": "8080",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{
						"foo": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:8080",
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
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{
				ExposedByDefault: true,
				DefaultRule:      "Host(`{{ normalize .Name }}.traefik.wtf`)",
			}
			p.Constraints = test.constraints

			err := p.Init()
			require.NoError(t, err)

			for i := 0; i < len(test.containers); i++ {
				var err error
				test.containers[i].ExtraConf, err = p.getConfiguration(test.containers[i])
				require.NoError(t, err)
			}

			configuration := p.buildConfiguration(context.Background(), test.containers)

			assert.Equal(t, test.expected, configuration)
		})
	}
}

func TestDockerGetIPPort(t *testing.T) {
	type expected struct {
		ip    string
		port  string
		error bool
	}

	testCases := []struct {
		desc       string
		container  docker.ContainerJSON
		serverPort string
		expected   expected
	}{
		{
			desc: "label traefik.port not set, no binding, falling back on the container's IP/Port",
			container: containerJSON(
				ports(nat.PortMap{
					"8080/tcp": {},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			expected: expected{
				ip:   "10.11.12.13",
				port: "8080",
			},
		},
		{
			desc: "label traefik.port not set, single binding with port only, falling back on the container's IP/Port",
			container: containerJSON(
				withNetwork("testnet", ipv4("10.11.12.13")),
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostPort: "8082",
						},
					},
				}),
			),
			expected: expected{
				ip:   "10.11.12.13",
				port: "80",
			},
		},
		{
			desc: "label traefik.port not set, binding with ip:port should create a route to the bound ip:port",
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1.2.3.4",
							HostPort: "8081",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			expected: expected{
				ip:   "1.2.3.4",
				port: "8081",
			},
		},
		{
			desc:       "label traefik.port set, no binding, falling back on the container's IP/traefik.port",
			container:  containerJSON(withNetwork("testnet", ipv4("10.11.12.13"))),
			serverPort: "80",
			expected: expected{
				ip:   "10.11.12.13",
				port: "80",
			},
		},
		{
			desc: "label traefik.port set, single binding with ip:port for the label, creates the route",
			container: containerJSON(
				ports(nat.PortMap{
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			serverPort: "443",
			expected: expected{
				ip:   "5.6.7.8",
				port: "8082",
			},
		},
		{
			desc: "label traefik.port set, no binding on the corresponding port, falling back on the container's IP/label.port",
			container: containerJSON(
				ports(nat.PortMap{
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			serverPort: "80",
			expected: expected{
				ip:   "10.11.12.13",
				port: "80",
			},
		},
		{
			desc: "label traefik.port set, multiple bindings on different ports, uses the label to select the correct (first) binding",
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1.2.3.4",
							HostPort: "8081",
						},
					},
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			serverPort: "80",
			expected: expected{
				ip:   "1.2.3.4",
				port: "8081",
			},
		},
		{
			desc: "label traefik.port set, multiple bindings on different ports, uses the label to select the correct (second) binding",
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1.2.3.4",
							HostPort: "8081",
						},
					},
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			serverPort: "443",
			expected: expected{
				ip:   "5.6.7.8",
				port: "8082",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			provider := &Provider{
				Network:       "testnet",
				UseBindPortIP: true,
			}

			actualIP, actualPort, actualError := provider.getIPPort(context.Background(), dData, test.serverPort)
			if test.expected.error {
				require.Error(t, actualError)
			} else {
				require.NoError(t, actualError)
			}
			assert.Equal(t, test.expected.ip, actualIP)
			assert.Equal(t, test.expected.port, actualPort)
		})
	}
}

func TestDockerGetPort(t *testing.T) {
	testCases := []struct {
		desc       string
		container  docker.ContainerJSON
		serverPort string
		expected   string
	}{
		{
			desc:      "no binding, no server port label",
			container: containerJSON(name("foo")),
			expected:  "",
		},
		{
			desc: "binding, no server port label",
			container: containerJSON(ports(nat.PortMap{
				"80/tcp": {},
			})),
			expected: "80",
		},
		{
			desc: "binding, multiple ports, no server port label",
			container: containerJSON(ports(nat.PortMap{
				"80/tcp":  {},
				"443/tcp": {},
			})),
			expected: "80",
		},
		{
			desc:       "no binding, server port label",
			container:  containerJSON(),
			serverPort: "8080",
			expected:   "8080",
		},
		{
			desc: "binding, server port label",
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": {},
				})),
			serverPort: "8080",
			expected:   "8080",
		},
		{
			desc: "binding, multiple ports, server port label",
			container: containerJSON(ports(nat.PortMap{
				"8080/tcp": {},
				"80/tcp":   {},
			})),
			serverPort: "8080",
			expected:   "8080",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getPort(dData, test.serverPort)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetIPAddress(t *testing.T) {
	testCases := []struct {
		desc      string
		container docker.ContainerJSON
		network   string
		expected  string
	}{
		{
			desc:      "one network, no network label",
			container: containerJSON(withNetwork("testnet", ipv4("10.11.12.13"))),
			expected:  "10.11.12.13",
		},
		{
			desc: "one network, network label",
			container: containerJSON(
				withNetwork("testnet", ipv4("10.11.12.13")),
			),
			network:  "testnet",
			expected: "10.11.12.13",
		},
		{
			desc: "two networks, network label",
			container: containerJSON(
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("testnet2", ipv4("10.11.12.14")),
			),
			network:  "testnet2",
			expected: "10.11.12.14",
		},
		{
			desc: "two networks, no network label, mode host",
			container: containerJSON(
				networkMode("host"),
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("testnet2", ipv4("10.11.12.14")),
			),
			expected: "127.0.0.1",
		},
		{
			desc: "two networks, no network label, mode host, use provider network",
			container: containerJSON(
				networkMode("host"),
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("webnet", ipv4("10.11.12.14")),
			),
			expected: "10.11.12.14",
		},
		{
			desc: "two networks, network label",
			container: containerJSON(
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("webnet", ipv4("10.11.12.14")),
			),
			network:  "testnet",
			expected: "10.11.12.13",
		},
		{
			desc: "no network, no network label, mode host",
			container: containerJSON(
				networkMode("host"),
			),
			expected: "127.0.0.1",
		},
		{
			desc: "no network, no network label, mode host, node IP",
			container: containerJSON(
				networkMode("host"),
				nodeIP("10.0.0.5"),
			),
			expected: "10.0.0.5",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			provider := &Provider{
				Network: "webnet",
			}

			dData := parseContainer(test.container)

			dData.ExtraConf.Docker.Network = provider.Network
			if len(test.network) > 0 {
				dData.ExtraConf.Docker.Network = test.network
			}

			actual := provider.getIPAddress(context.Background(), dData)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSwarmGetIPAddress(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(withEndpointSpec(modeDNSSR)),
			expected: "",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(
				withEndpointSpec(modeVIP),
				withEndpoint(virtualIP("1", "10.11.12.13/24")),
			),
			expected: "10.11.12.13",
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			service: swarmService(
				serviceLabels(map[string]string{
					"traefik.docker.network": "barnet",
				}),
				withEndpointSpec(modeVIP),
				withEndpoint(
					virtualIP("1", "10.11.12.13/24"),
					virtualIP("2", "10.11.12.99/24"),
				),
			),
			expected: "10.11.12.99",
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foonet",
				},
				"2": {
					Name: "barnet",
				},
			},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			provider := &Provider{
				SwarmMode: true,
			}

			dData, err := provider.parseService(context.Background(), test.service, test.networks)
			require.NoError(t, err)

			actual := provider.getIPAddress(context.Background(), dData)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSwarmGetPort(t *testing.T) {
	testCases := []struct {
		service    swarm.Service
		serverPort string
		networks   map[string]*docker.NetworkResource
		expected   string
	}{
		{
			service: swarmService(
				withEndpointSpec(modeDNSSR),
			),
			networks:   map[string]*docker.NetworkResource{},
			serverPort: "8080",
			expected:   "8080",
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			p := Provider{}

			dData, err := p.parseService(context.Background(), test.service, test.networks)
			require.NoError(t, err)

			actual := getPort(dData, test.serverPort)
			assert.Equal(t, test.expected, actual)
		})
	}
}
