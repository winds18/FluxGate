package stats

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultV2RayStatsPattern = "traffic"

type V2RayGRPCCollector struct {
	Addr           string
	QueryPattern   string
	Reset          bool
	DialTimeout    time.Duration
	RequestTimeout time.Duration
	DialOptions    []grpc.DialOption

	client v2RayStatsClient
}

func (c V2RayGRPCCollector) CollectCounters(ctx context.Context) ([]Counter, error) {
	client := c.client
	var closeConn func() error
	if client == nil {
		if strings.TrimSpace(c.Addr) == "" {
			return nil, errors.New("sing-box v2ray api address is required")
		}
		conn, err := c.dial(ctx)
		if err != nil {
			return nil, err
		}
		closeConn = conn.Close
		client = newV2RayStatsClient(conn)
	}
	if closeConn != nil {
		defer closeConn()
	}

	requestCtx, cancel := context.WithTimeout(ctx, c.requestTimeout())
	defer cancel()

	resp, err := client.QueryStats(requestCtx, &v2RayQueryStatsRequest{
		Pattern: c.queryPattern(),
		Reset_:  c.Reset,
	})
	if err != nil {
		return nil, fmt.Errorf("query sing-box v2ray stats: %w", err)
	}
	if resp == nil {
		return nil, nil
	}

	counters := make([]Counter, 0, len(resp.Stat))
	for _, stat := range resp.Stat {
		if stat == nil || strings.TrimSpace(stat.Name) == "" {
			continue
		}
		counters = append(counters, Counter{
			Name:  stat.Name,
			Value: stat.Value,
		})
	}
	return counters, nil
}

func (c V2RayGRPCCollector) dial(ctx context.Context) (*grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, c.dialTimeout())
	defer cancel()

	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}
	options = append(options, c.DialOptions...)
	conn, err := grpc.DialContext(dialCtx, c.Addr, options...)
	if err != nil {
		return nil, fmt.Errorf("dial sing-box v2ray api: %w", err)
	}
	return conn, nil
}

func (c V2RayGRPCCollector) queryPattern() string {
	if value := strings.TrimSpace(c.QueryPattern); value != "" {
		return value
	}
	return defaultV2RayStatsPattern
}

func (c V2RayGRPCCollector) dialTimeout() time.Duration {
	if c.DialTimeout > 0 {
		return c.DialTimeout
	}
	return 5 * time.Second
}

func (c V2RayGRPCCollector) requestTimeout() time.Duration {
	if c.RequestTimeout > 0 {
		return c.RequestTimeout
	}
	return 5 * time.Second
}

type v2RayStatsClient interface {
	QueryStats(ctx context.Context, in *v2RayQueryStatsRequest, opts ...grpc.CallOption) (*v2RayQueryStatsResponse, error)
}

type v2RayStatsGRPCClient struct {
	cc grpc.ClientConnInterface
}

func newV2RayStatsClient(cc grpc.ClientConnInterface) v2RayStatsClient {
	return v2RayStatsGRPCClient{cc: cc}
}

func (c v2RayStatsGRPCClient) QueryStats(ctx context.Context, in *v2RayQueryStatsRequest, opts ...grpc.CallOption) (*v2RayQueryStatsResponse, error) {
	out := new(v2RayQueryStatsResponse)
	err := c.cc.Invoke(ctx, "/v2ray.core.app.stats.command.StatsService/QueryStats", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type v2RayQueryStatsRequest struct {
	Pattern string `protobuf:"bytes,1,opt,name=pattern,proto3" json:"pattern,omitempty"`
	Reset_  bool   `protobuf:"varint,2,opt,name=reset,proto3" json:"reset,omitempty"`
}

func (m *v2RayQueryStatsRequest) Reset()         { *m = v2RayQueryStatsRequest{} }
func (m *v2RayQueryStatsRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*v2RayQueryStatsRequest) ProtoMessage()    {}

type v2RayQueryStatsResponse struct {
	Stat []*v2RayStat `protobuf:"bytes,1,rep,name=stat,proto3" json:"stat,omitempty"`
}

func (m *v2RayQueryStatsResponse) Reset()         { *m = v2RayQueryStatsResponse{} }
func (m *v2RayQueryStatsResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*v2RayQueryStatsResponse) ProtoMessage()    {}

type v2RayStat struct {
	Name  string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Value int64  `protobuf:"varint,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *v2RayStat) Reset()         { *m = v2RayStat{} }
func (m *v2RayStat) String() string { return fmt.Sprintf("%+v", *m) }
func (*v2RayStat) ProtoMessage()    {}
