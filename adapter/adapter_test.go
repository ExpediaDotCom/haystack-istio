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

package haystack

import (
	"context"
	"log"
	"testing"
	"time"

	client "github.com/ExpediaDotCom/haystack-client-go"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	istio_mixer_v1 "istio.io/api/mixer/v1"
	"istio.io/istio/mixer/pkg/attribute"
)

func TestReport(t *testing.T) {
	conn, err := grpc.Dial("mixs:9091", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Unable to connect to gRPC server: %v", err)
	}
	client := istio_mixer_v1.NewMixerClient(conn)

	// send one server and client span
	for _, req := range []*istio_mixer_v1.ReportRequest{createClientSpan(), createServerSpan()} {
		_, rptErr := client.Report(context.Background(), req)
		if rptErr != nil {
			t.Fatalf("Unable to connect to gRPC server: %v", err)
		}
	}

	verifyFromKafkaReads(t)
}

func verifyFromKafkaReads(t *testing.T) {
	consumer, err := sarama.NewConsumer([]string{"kafkasvc:9092"}, nil)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition("proto-spans", 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			panic(err)
		}
	}()

	clientSpanReceived := 0
	serverSpanReceived := 0

ConsumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("Consumed message offset %d\n", msg.Offset)
			span := &client.Span{}
			unmarshalErr := span.XXX_Unmarshal(msg.Value)
			if unmarshalErr != nil {
				panic(err)
			}

			for _, tag := range span.GetTags() {
				if tag.GetKey() == "span.kind" {
					switch tag.GetVStr() {
					case "client":
						clientSpanReceived = clientSpanReceived + 1
						verifyClientSpan(t, span)
					case "server":
						serverSpanReceived = serverSpanReceived + 1
						verifyServerSpan(t, span)
					}
				}
			}

			// expect only two spans
			if msg.Offset == 1 {
				assert.Equal(t, clientSpanReceived, 1)
				assert.Equal(t, serverSpanReceived, 1)
				break ConsumerLoop
			}
		}
	}
}

func createServerSpan() *istio_mixer_v1.ReportRequest {
	httpHeaders := map[string]interface{}{
		"x-b3-traceid":      "sTraceid",
		"x-b3-spanid":       "sSpanid",
		"x-b3-parentspanid": "sParentid",
	}

	now := time.Now()

	attrs := map[string]interface{}{
		"request.size":          555,
		"response.size":         100,
		"destination.service":   "dest",
		"context.reporter.kind": "inbound",
		"request.time":          now,
		"response.time":         now.Add(50 * time.Millisecond),
		"request.headers":       httpHeaders,
	}

	return &istio_mixer_v1.ReportRequest{
		Attributes: []istio_mixer_v1.CompressedAttributes{
			getAttrBag(attrs)},
	}
}

func createClientSpan() *istio_mixer_v1.ReportRequest {
	httpHeaders := map[string]interface{}{
		"x-b3-traceid":      "cTraceid",
		"x-b3-spanid":       "cSpanid",
		"x-b3-parentspanid": "cParentid",
	}

	now := time.Now()
	attrs := map[string]interface{}{
		"request.size":          600,
		"response.size":         200,
		"destination.service":   "dest",
		"context.reporter.kind": "outbound",
		"response.time":         now.Add(50 * time.Millisecond),
		"request.time":          now,
		"request.headers":       httpHeaders,
	}

	return &istio_mixer_v1.ReportRequest{
		Attributes: []istio_mixer_v1.CompressedAttributes{
			getAttrBag(attrs)},
	}
}

func verifyClientSpan(t *testing.T, span *client.Span) {
	assert.Equal(t, span.TraceId, "cTraceid")
	assert.Equal(t, span.SpanId, "cSpanid")
	assert.Equal(t, span.ParentSpanId, "cParentid")
	assert.Equal(t, span.Duration, int64(50000))
	assert.WithinDuration(t, time.Now(), time.Unix(span.StartTime/1000000, 0), 10*time.Second)
}

func verifyServerSpan(t *testing.T, span *client.Span) {
	assert.Equal(t, span.TraceId, "sTraceid")
	assert.Equal(t, span.SpanId, "sSpanid")
	assert.Equal(t, span.ParentSpanId, "sParentid")
	assert.Equal(t, span.Duration, int64(50000))
	assert.WithinDuration(t, time.Now(), time.Unix(span.StartTime/1000000, 0), 10*time.Second)
}

func getAttrBag(attrs map[string]interface{}) istio_mixer_v1.CompressedAttributes {
	requestBag := attribute.GetMutableBag(nil)
	for k, v := range attrs {
		switch v.(type) {
		case map[string]interface{}:
			mapCast := make(map[string]string, len(v.(map[string]interface{})))

			for k1, v1 := range v.(map[string]interface{}) {
				mapCast[k1] = v1.(string)
			}
			requestBag.Set(k, mapCast)
		default:
			requestBag.Set(k, v)
		}
	}

	var attrProto istio_mixer_v1.CompressedAttributes
	requestBag.ToProto(&attrProto, nil, 0)
	return attrProto
}
