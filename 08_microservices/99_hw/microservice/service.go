package main

import (
	context "context"
	"encoding/json"
	"log"
	"net"

	"google.golang.org/grpc"
)

func parseACL(acl string) (map[string][]string, error) {
	var mapACL map[string][]string

	err := json.Unmarshal([]byte(acl), &mapACL)
	if err != nil {
		return nil, err
	}

	return mapACL, nil
}

func StartMyMicroservice(ctx context.Context, addr string, acl string) error {
	mapACL, err := parseACL(acl)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("can't listen this address: %s, err: %v", addr, err)
	}

	admin := NewAdminService()

	wrapper := &serverWrapper{acl: mapACL, adminService: admin}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(wrapper.authInterceptor),
		grpc.StreamInterceptor(wrapper.authStreamInterceptor),
	)

	RegisterAdminServer(server, admin)
	RegisterBizServer(server, &BizService{})

	go func() {
		err := server.Serve(lis)
		if err != nil {
			log.Fatalf("can't start server: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	return nil
}
