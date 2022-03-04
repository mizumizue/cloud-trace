package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
)

func main() {
	err := InitTracerProvider(1)
	if err != nil {
		log.Fatalln(err)
	}
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer func(tp *sdktrace.TracerProvider, ctx context.Context) {
		// flushes any pending spans
		err := tp.ForceFlush(ctx)
		if err != nil {
			log.Println(err)
		}
	}(tp, ctx)

	err := WithTrace(ctx, "hoge.com/trace", "rootspanfoo", func(ctx context.Context) error {
		log.Println("hoge")
		time.Sleep(time.Millisecond * 100)
		log.Println("foo")
		time.Sleep(time.Millisecond * 100)
		log.Println("bar")
		for i := 0; i < 10; i++ {
			_ = WithSpan(ctx, fmt.Sprintf("childspan%d", i+1), func(ctx context.Context) error {
				log.Println(i)
				time.Sleep(time.Millisecond * 100)
				return nil
			})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

var tp *sdktrace.TracerProvider

// InitTracerProvider Trace用のTracerProviderを初期化します。fractionを1以上にすることで全てのRequestをSamplingします
func InitTracerProvider(fraction float64) error {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		return fmt.Errorf("texporter.NewExporter: %v", err)
	}
	if tp == nil {
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithSampler(
				sdktrace.TraceIDRatioBased(fraction), // TODO 分数で設定する。本番のRequest数を元に決めたい
			),
		)
		otel.SetTracerProvider(tp)
	}
	return nil
}

func WithTrace(ctx context.Context, instrumentationName, spanName string, f func(ctx context.Context) error) error {
	defer func(tp *sdktrace.TracerProvider, ctx context.Context) {
		err := tp.ForceFlush(ctx) // flushes any pending spans
		if err != nil {
			log.Println(err)
		}
	}(tp, ctx)
	ctx = context.WithValue(ctx, "INSTRUMENTATION_NAME", instrumentationName)
	tracer := otel.Tracer(instrumentationName)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()
	return f(ctx)
}

func WithSpan(ctx context.Context, spanName string, f func(ctx context.Context) error) error {
	in := ctx.Value("INSTRUMENTATION_NAME").(string)
	ctx, span := otel.Tracer(in).Start(ctx, spanName)
	defer span.End()
	return f(ctx)
}
