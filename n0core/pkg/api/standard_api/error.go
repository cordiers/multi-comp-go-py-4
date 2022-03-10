package stdapi

import (
	"google.golang.org/grpc/codes"

	grpcutil "github.com/n0stack/n0stack/n0core/pkg/util/grpc"
)

func LockError() error {
	return grpcutil.WrapGrpcErrorf(codes.FailedPrecondition, "this is locked, wait a moment")
}
