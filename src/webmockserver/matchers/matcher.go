package matchers

import (
	"net/http"
	"sync"
	proto "webmockserver/proto"
)

type Matcher struct {
	assertions         map[*Assertion][]*http.Request
	originalAssertions *Assertions
	mutex              sync.Mutex
}

func NewMatcher() *Matcher {
	return new(Matcher)
}

func (matcher *Matcher) Set(assertions *Assertions) {
	matcher.mutex.Lock()
	defer matcher.mutex.Unlock()
	matcher.assertions = make(map[*Assertion][]*http.Request)
	matcher.originalAssertions = assertions
	for _, assertion := range assertions.assertions {
		matcher.assertions[assertion] = make([]*http.Request, 0, assertion.atLeast)
	}
}

func (matcher *Matcher) Get() *Assertions {
	matcher.mutex.Lock()
	defer matcher.mutex.Unlock()
	return matcher.originalAssertions
}

func (matcher *Matcher) MatchRequest(request *http.Request) *Assertion {
	matcher.mutex.Lock()
	defer matcher.mutex.Unlock()

	defer request.Body.Close()

	for assertion, requests := range matcher.assertions {
		if err := assertion.MatchRequest(request); err == nil {
			matcher.assertions[assertion] = append(requests, request)
			return assertion
		}
	}

	return nil
}

func (_ *Matcher) WriteToResponse(assertion *Assertion, response http.ResponseWriter) error {
	for key, values := range assertion.returnHeaders {
		for _, value := range values {
			response.Header().Add(key, value)
		}
	}
	var code int = 200
	if assertion.returnStatusCode != 0 {
		code = int(assertion.returnStatusCode)
	}
	response.WriteHeader(code)
	_, err := response.Write(assertion.returnBody)
	return err
}

func (matcher *Matcher) ExportToProtobuffer() *proto.WebMockResults {
	matcher.mutex.Lock()
	defer matcher.mutex.Unlock()

	results := new(proto.WebMockResults)
	results.Success = true

	for assertion, requests := range matcher.assertions {
		result := new(proto.WebMockResult)
		result.Success = true
		result.Assertion = assertion.ToProtoBuffer()
		result.MatchRequests = make([]*proto.MatchRequest, 0, len(requests))
		for _, request := range requests {
			result.MatchRequests = append(result.MatchRequests, convertHttpRequestToProtobuffer(request))
		}

		if assertion.atLeast > int64(len(requests)) || assertion.atMost < int64(len(requests)) {
			result.Success = false
			results.Success = false
		}
		results.AssertionResults = append(results.AssertionResults, result)
	}

	return results
}

func (matcher *Matcher) Clear() {
	matcher.mutex.Lock()
	defer matcher.mutex.Unlock()
	matcher.assertions = nil
	matcher.originalAssertions = nil
}

func convertHttpRequestToProtobuffer(request *http.Request) *proto.MatchRequest {
	return &proto.MatchRequest{
		Host:       request.Host,
		Method:     request.Method,
		Path:       request.URL.Path,
		Headers:    convertMapFromGolangToProto(request.Header),
		Parameters: convertMapFromGolangToProto(request.URL.Query()),
	}
}
