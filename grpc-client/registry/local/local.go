/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/11/18
   Description :
-------------------------------------------------
*/

package local

import (
	"errors"
	"strings"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"

	"github.com/zly-app/zapp/logger"
)

const Name = "local"

var defaultResolver = newResolver()

func RegistryAddress(endpointName, address string) {
	defaultResolver.RegistryEndpoint(endpointName, address)
}

type resolverCli struct {
	endpoints map[string][]resolver.Address
	once      sync.Once
}

func newResolver() *resolverCli {
	return &resolverCli{
		endpoints: make(map[string][]resolver.Address),
	}
}

func (r *resolverCli) RegistryEndpoint(endpointName, endpoints string) {
	if endpoints == "" {
		logger.Log.Fatal("endpoint is empty", zap.String("name", endpointName))
	}
	address := strings.Split(endpoints, ",")
	addr := make([]resolver.Address, len(address))
	for i, a := range address {
		addr[i] = resolver.Address{Addr: a}
	}

	r.endpoints[endpointName] = addr

	r.once.Do(func() {
		resolver.Register(defaultResolver)
	})
}

func (r *resolverCli) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	address := r.endpoints[target.Endpoint]
	if len(address) == 0 {
		return nil, errors.New("address of endpoint is empty or unregistered")
	}

	cc.UpdateState(resolver.State{Addresses: address})
	return r, nil
}
func (r *resolverCli) Scheme() string { return Name }

func (r *resolverCli) ResolveNow(options resolver.ResolveNowOptions) {}
func (r *resolverCli) Close()                                        {}
