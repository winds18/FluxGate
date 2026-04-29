package stats

import (
	"context"
	"errors"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type fakeV2RayStatsClient struct {
	request  *v2RayQueryStatsRequest
	response *v2RayQueryStatsResponse
	err      error
}

func (f *fakeV2RayStatsClient) QueryStats(_ context.Context, in *v2RayQueryStatsRequest, _ ...grpc.CallOption) (*v2RayQueryStatsResponse, error) {
	f.request = in
	if f.err != nil {
		return nil, f.err
	}
	return f.response, nil
}

func TestV2RayGRPCCollectorCollectsCounters(t *testing.T) {
	client := &fakeV2RayStatsClient{
		response: &v2RayQueryStatsResponse{Stat: []*v2RayStat{
			{Name: "user>>>fg_u_1_t_1>>>traffic>>>uplink", Value: 120},
			{Name: "  ", Value: 999},
			nil,
			{Name: "outbound>>>up_42>>>traffic>>>downlink", Value: 340},
		}},
	}
	collector := V2RayGRPCCollector{
		QueryPattern: "traffic",
		client:       client,
	}

	counters, err := collector.CollectCounters(context.Background())
	if err != nil {
		t.Fatalf("collect counters: %v", err)
	}
	if client.request == nil || client.request.Pattern != "traffic" || client.request.Reset_ {
		t.Fatalf("unexpected query request: %+v", client.request)
	}
	if len(counters) != 2 {
		t.Fatalf("expected two counters, got %+v", counters)
	}
	if counters[0].Name != "user>>>fg_u_1_t_1>>>traffic>>>uplink" || counters[0].Value != 120 {
		t.Fatalf("unexpected first counter: %+v", counters[0])
	}
	if counters[1].Name != "outbound>>>up_42>>>traffic>>>downlink" || counters[1].Value != 340 {
		t.Fatalf("unexpected second counter: %+v", counters[1])
	}
}

func TestV2RayGRPCCollectorUsesDefaultPattern(t *testing.T) {
	client := &fakeV2RayStatsClient{response: &v2RayQueryStatsResponse{}}
	collector := V2RayGRPCCollector{client: client}

	if _, err := collector.CollectCounters(context.Background()); err != nil {
		t.Fatalf("collect counters: %v", err)
	}
	if client.request == nil || client.request.Pattern != defaultV2RayStatsPattern {
		t.Fatalf("unexpected default query request: %+v", client.request)
	}
}

func TestV2RayGRPCCollectorReturnsQueryError(t *testing.T) {
	wantErr := errors.New("query failed")
	collector := V2RayGRPCCollector{client: &fakeV2RayStatsClient{err: wantErr}}

	_, err := collector.CollectCounters(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped query error, got %v", err)
	}
}

func TestV2RayGRPCCollectorQueriesGRPCService(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	service := &testStatsService{
		response: &v2RayQueryStatsResponse{Stat: []*v2RayStat{
			{Name: "user>>>fg_u_1_t_1>>>traffic>>>uplink", Value: 10},
		}},
	}
	server.RegisterService(&testStatsServiceDesc, service)
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	collector := V2RayGRPCCollector{
		Addr:         "bufnet",
		QueryPattern: "traffic",
		DialOptions: []grpc.DialOption{
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return listener.Dial()
			}),
		},
	}

	counters, err := collector.CollectCounters(context.Background())
	if err != nil {
		t.Fatalf("collect counters over grpc: %v", err)
	}
	if service.request == nil || service.request.Pattern != "traffic" {
		t.Fatalf("unexpected service request: %+v", service.request)
	}
	if len(counters) != 1 || counters[0].Value != 10 {
		t.Fatalf("unexpected grpc counters: %+v", counters)
	}
}

type testStatsService struct {
	request  *v2RayQueryStatsRequest
	response *v2RayQueryStatsResponse
}

func (s *testStatsService) QueryStats(_ context.Context, in *v2RayQueryStatsRequest) (*v2RayQueryStatsResponse, error) {
	s.request = in
	return s.response, nil
}

type testStatsServiceServer interface {
	QueryStats(context.Context, *v2RayQueryStatsRequest) (*v2RayQueryStatsResponse, error)
}

var testStatsServiceDesc = grpc.ServiceDesc{
	ServiceName: "v2ray.core.app.stats.command.StatsService",
	HandlerType: (*testStatsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "QueryStats",
			Handler:    testStatsQueryStatsHandler,
		},
	},
}

func testStatsQueryStatsHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(v2RayQueryStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(testStatsServiceServer).QueryStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v2ray.core.app.stats.command.StatsService/QueryStats",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(testStatsServiceServer).QueryStats(ctx, req.(*v2RayQueryStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}
