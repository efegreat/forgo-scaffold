package support

import (
	"testing"

	"github.com/efegreat/forgo/forgotest"
	fgrpc "github.com/efegreat/forgo/server/grpc"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	orderv1 "example/api/order/v1"
)

// Conn returns a connection to env's gRPC server over loopback, which closes when t ends. grpc-go serves
// calls only through its transport, so gRPC tests go over the network, as clients do.
func Conn(t testing.TB, env *forgotest.Env) *grpc.ClientConn {
	t.Helper()
	addr, err := env.Addr(fgrpc.ServerNameGRPC)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// Orders returns a client of the orders API of env's gRPC server, the one other services use.
func Orders(t testing.TB, env *forgotest.Env) orderv1.OrderServiceClient {
	t.Helper()
	return orderv1.NewOrderServiceClient(Conn(t, env))
}

// Status returns the code of err, the error of a gRPC call, and its error code, such as "validation": the
// reason of its ErrorInfo, which the server sends with every error. It returns codes.OK for a nil err.
func Status(err error) (codes.Code, string) {
	st := status.Convert(err)
	for _, detail := range st.Details() {
		if info, ok := detail.(*errdetails.ErrorInfo); ok {
			return st.Code(), info.GetReason()
		}
	}
	return st.Code(), ""
}
