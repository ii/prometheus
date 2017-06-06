package apiv2

import (
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/tsdb"
)

type API struct {
	db *tsdb.DB
}

func New(db *tsdb.DB) *API {
	return &API{
		db: db,
	}
}

func (api *API) RegisterGRPC(srv *grpc.Server) {
	RegisterTSDBAdminServer(srv, &tsdbAdminService{db: api.db})
}

func (api *API) HTTPHandler(grpcAddr string) (http.Handler, error) {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := RegisterTSDBAdminHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
	if err != nil {
		return nil, err
	}
	return mux, nil
}

type tsdbAdminService struct {
	db *tsdb.DB
}

func (s *tsdbAdminService) Reload(ctx context.Context, _ *Empty) (*Empty, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *tsdbAdminService) Snapshot(ctx context.Context, _ *TSDBAdminSnapshotRequest) (*TSDBAdminSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *tsdbAdminService) DeleteSeries(ctx context.Context, _ *TSDBAdminDeleteRequest) (*TSDBAdminDeleteResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
