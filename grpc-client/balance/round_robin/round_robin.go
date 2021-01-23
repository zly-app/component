/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/11/19
   Description :
-------------------------------------------------
*/

package round_robin

import (
	"google.golang.org/grpc"
)

const Name = "round_robin"

func Balance() grpc.DialOption {
	return grpc.WithDefaultServiceConfig(`{ "loadBalancingConfig": [{"round_robin": {}}] }`)
}
