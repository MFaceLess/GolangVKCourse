package main

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type serverWrapper struct {
	acl          map[string][]string
	adminService *AdminService
}

func (sw *serverWrapper) authInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if err := sw.checkAccess(ctx, info.FullMethod); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (sw *serverWrapper) authStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	if err := sw.checkAccess(ss.Context(), info.FullMethod); err != nil {
		return err
	}
	return handler(srv, ss)
}

func (sw *serverWrapper) checkAccess(ctx context.Context, method string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "access denied")
	}

	consumers := md["consumer"]
	if len(consumers) == 0 {
		return status.Error(codes.Unauthenticated, "access denied")
	}

	consumer := consumers[0]

	allowedMethods, ok := sw.acl[consumer]
	if !ok {
		return status.Error(codes.Unauthenticated, "access denied")
	}

	var host string
	if p, ok := peer.FromContext(ctx); ok {
		host = p.Addr.String()
	}

	for _, m := range allowedMethods {
		if m == method || (strings.HasSuffix(m, "*") && strings.HasPrefix(method, strings.TrimSuffix(m, "*"))) {
			event := &Event{
				Timestamp: time.Now().Unix(),
				Consumer:  consumer,
				Method:    method,
				Host:      host,
			}
			sw.adminService.broadcast(event)
			return nil
		}
	}

	return status.Error(codes.Unauthenticated, "access denied")
}
