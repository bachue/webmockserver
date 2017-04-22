package main

import (
	"net/http"
	"webmockserver/matchers"
	proto "webmockserver/proto"

	"golang.org/x/net/context"
)

type Server struct {
	matcher *matchers.Matcher
}

func NewServer() *Server {
	return &Server{
		matcher: matchers.NewMatcher(),
	}
}

func (server *Server) Set(ctx context.Context, protoAssertions *proto.WebMockAssertions) (*proto.Void, error) {
	assertions := matchers.NewAssertionsFromProtobuf(protoAssertions)
	server.matcher.Set(assertions)
	return new(proto.Void), nil
}

func (server *Server) Get(ctx context.Context, _ *proto.Void) (*proto.WebMockAssertions, error) {
	assertions := server.matcher.Get()
	if assertions == nil {
		return nil, nil
	}
	return assertions.ToProtoBuffer(), nil
}

func (server *Server) Done(ctx context.Context, _ *proto.Void) (*proto.WebMockResults, error) {
	results := server.matcher.ExportToProtobuffer()
	server.matcher.Clear()
	return results, nil
}

func (server *Server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if assertion := server.matcher.MatchRequest(request); assertion != nil {
		server.matcher.WriteToResponse(assertion, response)
	} else {
		response.WriteHeader(404)
	}
}
