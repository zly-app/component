/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/11/18
   Description :
-------------------------------------------------
*/

package static

import (
	"errors"
	"strings"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"

	"github.com/zly-app/zapp/logger"
)

const Name = "static"

var defaultResolver = newResolver()

func RegistryAddress(serviceName, address string) {
	defaultResolver.RegistryEndpoint(serviceName, address)
}

type resolverCli struct {
	endpoints map[string][]resolver.Address
	mx        sync.RWMutex
}

func newResolver() *resolverCli {
	r := &resolverCli{
		endpoints: make(map[string][]resolver.Address),
	}
	resolver.Register(r)
	return r
}

func (r *resolverCli) RegistryEndpoint(serviceName, address string) {
	if address == "" {
		logger.Log.Fatal("endpoint is empty", zap.String("name", serviceName))
	}

	ss := strings.Split(address, ",")
	addresses := make([]resolver.Address, len(ss))
	for i, s := range ss {
		addresses[i] = resolver.Address{Addr: s}
	}

	r.mx.Lock()
	r.endpoints[serviceName] = addresses
	r.mx.Unlock()
}

func (r *resolverCli) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.mx.RLock()
	address := r.endpoints[target.Endpoint]
	r.mx.RUnlock()
	if len(address) == 0 {
		return nil, errors.New("address of endpoint is empty or unregistered")
	}

	cc.UpdateState(resolver.State{Addresses: address})
	return r, nil
}
func (r *resolverCli) Scheme() string { return Name }

func (r *resolverCli) ResolveNow(options resolver.ResolveNowOptions) {}
func (r *resolverCli) Close()                                        {}
