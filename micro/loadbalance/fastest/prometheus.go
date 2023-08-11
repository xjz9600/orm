package fastest

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"time"
)

type PrometheusBuilder struct {
	Name       string
	Port       string
	Help       string
	ServerName string
}

func NewPrometheusBuilder(serverName, name, port, help string) *PrometheusBuilder {
	return &PrometheusBuilder{
		Name:       name,
		Port:       port,
		Help:       help,
		ServerName: serverName,
	}
}

//func getIP() string {
//	conn, err := net.Dial("udp", "8.8.8.8:80")
//	if err != nil {
//		return ""
//	}
//	defer conn.Close()
//
//	localAddr := conn.LocalAddr().(*net.UDPAddr)
//	return localAddr.IP.String()
//}

func (p *PrometheusBuilder) BuildUnaryInterceptor() grpc.UnaryServerInterceptor {
	addr := "[::]"
	if p.Port != "" {
		addr += p.Port
	} else {
		addr += ":80"
	}
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: p.Name,
		Help: p.Help,
		Objectives: map[float64]float64{
			0.5: 0.01,
		},
		ConstLabels: map[string]string{
			"addr": addr,
		},
	}, []string{"kind"})
	prometheus.MustRegister(vector)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()
		defer func() {
			duration := time.Now().Sub(startTime).Nanoseconds()
			vector.WithLabelValues(p.ServerName).Observe(float64(duration))
		}()
		return handler(ctx, req)
	}
}
