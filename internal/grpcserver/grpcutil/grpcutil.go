package grpcutil

import (
	"strings"

	"google.golang.org/grpc/metadata"
)

func GetMDString(md metadata.MD, key string) string {
	v := md.Get(strings.ToLower(key))
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func ParseBearer(authz string) (string, bool) {
	authz = strings.TrimSpace(authz)
	if authz == "" {
		return "", false
	}
	const p = "bearer "
	if len(authz) < len(p) || strings.ToLower(authz[:len(p)]) != p {
		return "", false
	}
	tok := strings.TrimSpace(authz[len(p):])
	return tok, tok != ""
}