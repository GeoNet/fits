// token allows for token based basic auth with grpc.
package token

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
)

var secure = true

type token struct {
	t string
}

func New(t string) credentials.PerRPCCredentials {
	return token{t: t}
}

func (t token) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"token": t.t,
	}, nil
}

func (t token) RequireTransportSecurity() bool {
	return secure
}
