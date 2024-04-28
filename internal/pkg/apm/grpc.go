package apm

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

type grpcHandler struct {
	stats.Handler
	customHandler func(ctx context.Context, rs stats.RPCStats)
}

func (gh *grpcHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	gh.customHandler(ctx, rs)
}

func OtelGRPCNewServerHandler(ignoredMethods ...string) stats.Handler {
	checkList := map[string]struct{}{}
	for _, m := range ignoredMethods {
		checkList[m] = struct{}{}
	}

	handler := otelgrpc.NewServerHandler(
		otelgrpc.WithTracerProvider(Global().GetTracerProvider()),
		otelgrpc.WithMeterProvider(Global().GetMeterProvider()),
	)

	gh := &grpcHandler{}
	gh.Handler = handler
	gh.customHandler = func(ctx context.Context, rs stats.RPCStats) {
		methodName, ok := grpc.Method(ctx)
		if !ok {
			return
		}

		_, skip := checkList[methodName]
		if !skip {
			return
		}

		handler.HandleRPC(ctx, rs)
	}

	return gh
}
