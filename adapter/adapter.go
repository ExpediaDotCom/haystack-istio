/*
 *  Copyright 2018 Expedia, Inc.
 *
 *     Licensed under the Apache License, Version 2.0 (the "License");
 *     you may not use this file except in compliance with the License.
 *     You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *     Unless required by applicable law or agreed to in writing, software
 *     distributed under the License is distributed on an "AS IS" BASIS,
 *     WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *     See the License for the specific language governing permissions and
 *     limitations under the License.
 *
 */

// nolint:lll
// Generates the haystack adapter's resource yaml. It contains the adapter's configuration, name,
// supported template names (tracespan in this case).
//go:generate $GOPATH/src/istio.io/istio/bin/mixer_codegen.sh -a mixer/adapter/haystack/config/config.proto -x "-s=false -n haystack -t tracespan"

package haystack

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"

	client "github.com/ExpediaDotCom/haystack-client-go"
	"github.com/gogo/protobuf/types"
	otext "github.com/opentracing/opentracing-go/ext"
	"istio.io/api/mixer/adapter/model/v1beta1"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/tracespan"
	"istio.io/istio/pkg/log"
)

type (
	// Server is basic server interface
	Server interface {
		Addr() string
		Close() error
		Run(shutdown chan error)
	}

	//GrpcAdapter supports tracing template.
	GrpcAdapter struct {
		listener   net.Listener
		server     *grpc.Server
		dispatcher client.Dispatcher
	}
)

var (
	clientKind = "client"
	serverKind = "server"
)

var _ tracespan.HandleTraceSpanServiceServer = &GrpcAdapter{}

// HandleTraceSpan records span
func (s *GrpcAdapter) HandleTraceSpan(ctx context.Context, r *tracespan.HandleTraceSpanRequest) (*v1beta1.ReportResult, error) {
	log.Debugf("received request %v\n", *r)

	for _, istioSpan := range r.Instances {
		span := s.convertIstioSpan(istioSpan)
		log.Infof("Dispatching the span to haystack-agent...")
		s.dispatcher.DispatchProtoSpan(span)
	}
	return &v1beta1.ReportResult{}, nil
}

func (s *GrpcAdapter) convertToMicros(t *types.Timestamp) int64 {
	return t.GetSeconds()*1000*1000 + int64(t.GetNanos()/1000)
}

func (s *GrpcAdapter) convertIstioSpan(istioSpan *tracespan.InstanceMsg) *client.Span {
	log.Debugf("Converting istio span to haystack proto span...")
	startTime := s.convertToMicros(istioSpan.StartTime.GetValue())
	endTime := s.convertToMicros(istioSpan.EndTime.GetValue())
	duration := endTime - startTime

	kind := &serverKind
	if istioSpan.ClientSpan {
		kind = &clientKind
	}

	tags := map[string]string{}
	tags["span.kind"] = *kind

	operationName := istioSpan.SpanName
	if strings.Contains(operationName, "?") {
		idx := strings.LastIndex(operationName, "?")
		qs := operationName[idx+1:]
		if vals, err := url.ParseQuery(qs); err == nil {
			for k, v := range vals {
				tags[k] = strings.Join(v, "; ")
			}
		}
		operationName = operationName[:idx]
	}

	if istioSpan.HttpStatusCode != 0 {
		tags[string(otext.HTTPStatusCode)] = strconv.FormatInt(istioSpan.HttpStatusCode, 10)
		if istioSpan.HttpStatusCode >= 400 {
			tags["error"] = "true"
		} else {
			tags["error"] = "false"
		}
	} else {
		tags["error"] = "false"
	}

	serviceName := ""
	if n, ok := istioSpan.SpanTags["source.app"]; ok {
		serviceName = n.GetStringValue()
	}

	for k, v := range istioSpan.SpanTags {
		shouldSet := k != "response.size" &&
			k != "request.size" &&
			k != "source.app"

		if s := adapter.Stringify(v); s != "" && shouldSet {
			tags[k] = v.GetStringValue()
		}
	}

	if v, ok := istioSpan.SpanTags["request.size"]; ok {
		tags["request.size"] = adapter.Stringify(v.GetInt64Value())
	}

	if v, ok := istioSpan.SpanTags["response.size"]; ok {
		tags["response.size"] = adapter.Stringify(v.GetInt64Value())
	}

	var protoTags []*client.Tag
	for k, v := range tags {
		protoTags = append(protoTags, client.ConvertToProtoTag(k, v))
	}

	return &client.Span{
		TraceId:       istioSpan.TraceId,
		SpanId:        istioSpan.SpanId,
		ParentSpanId:  istioSpan.ParentSpanId,
		ServiceName:   serviceName,
		OperationName: operationName,
		StartTime:     int64(startTime),
		Duration:      int64(duration),
		Tags:          protoTags,
	}
}

// Addr returns the listening address of the server
func (s *GrpcAdapter) Addr() string {
	return s.listener.Addr().String()
}

// Run starts the server run
func (s *GrpcAdapter) Run(shutdown chan error) {
	shutdown <- s.server.Serve(s.listener)
}

// Close gracefully shuts down the server; used for testing
func (s *GrpcAdapter) Close() error {
	if s.server != nil {
		s.server.GracefulStop()
	}

	if s.listener != nil {
		_ = s.listener.Close()
	}

	if s.dispatcher != nil {
		s.dispatcher.Close()
	}
	return nil
}

type consoleLogger struct{}

/*Error prints the error message*/
func (logger *consoleLogger) Error(format string, v ...interface{}) {
	log.Errorf(format, v)
}

/*Info prints the info message*/
func (logger *consoleLogger) Info(format string, v ...interface{}) {
	log.Infof(format, v)
}

/*Debug prints the info message*/
func (logger *consoleLogger) Debug(format string, v ...interface{}) {
	log.Debugf(format, v)
}

// NewHastackGrpcAdapter creates a new IBP adapter that listens at provided port.
func NewHastackGrpcAdapter(addr string, agentHost string, agentPort int) (Server, error) {
	if addr == "" {
		addr = "0"
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", addr))
	if err != nil {
		return nil, fmt.Errorf("unable to listen on socket: %v", err)
	}

	agentDispatcher := client.NewAgentDispatcher(agentHost, agentPort, 3*time.Second, 1000)
	agentDispatcher.SetLogger(&consoleLogger{})

	s := &GrpcAdapter{
		listener:   listener,
		dispatcher: agentDispatcher,
	}
	log.Infof("listening haystack grpc server on \"%v\"\n", s.Addr())
	s.server = grpc.NewServer()
	tracespan.RegisterHandleTraceSpanServiceServer(s.server, s)
	return s, nil
}
