package matchers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	proto "webmockserver/proto"
)

type Assertion struct {
	host             string
	method           string
	path             string
	headers          http.Header
	parameters       url.Values
	body             []byte
	matchHost        bool
	matchMethod      bool
	matchPath        bool
	matchHeaders     bool
	matchParameters  bool
	matchBody        bool
	atLeast          int64
	atMost           int64
	returnStatusCode int32
	returnHeaders    http.Header
	returnBody       []byte
}

func NewAssertionFromProtobuf(assertion *proto.WebMockAssertion) *Assertion {
	return &Assertion{
		host:             assertion.Host,
		method:           assertion.Method,
		path:             assertion.Path,
		headers:          http.Header(convertMapFromProtoToGolang(assertion.Headers)),
		parameters:       url.Values(convertMapFromProtoToGolang(assertion.Parameters)),
		body:             assertion.Body,
		matchHost:        assertion.MatchHost,
		matchMethod:      assertion.MatchMethod,
		matchPath:        assertion.MatchPath,
		matchHeaders:     assertion.MatchHeaders,
		matchParameters:  assertion.MatchParameters,
		matchBody:        assertion.MatchBody,
		atLeast:          assertion.AtLeast,
		atMost:           assertion.AtMost,
		returnStatusCode: assertion.ReturnStatusCode,
		returnHeaders:    http.Header(convertMapFromProtoToGolang(assertion.ReturnHeaders)),
		returnBody:       assertion.ReturnBody,
	}
}

func convertMapFromProtoToGolang(input map[string]*proto.StringArray) map[string][]string {
	output := make(map[string][]string)
	for key, array := range input {
		output[key] = array.Strings
	}
	return output
}

func convertMapFromGolangToProto(input map[string][]string) map[string]*proto.StringArray {
	output := make(map[string]*proto.StringArray)
	for key, array := range input {
		output[key] = &proto.StringArray{Strings: array}
	}
	return output
}

func (assertion *Assertion) ToProtoBuffer() *proto.WebMockAssertion {
	return &proto.WebMockAssertion{
		Host:             assertion.host,
		Method:           assertion.method,
		Path:             assertion.path,
		Headers:          convertMapFromGolangToProto(assertion.headers),
		Parameters:       convertMapFromGolangToProto(assertion.parameters),
		Body:             assertion.body,
		MatchHost:        assertion.matchHost,
		MatchMethod:      assertion.matchMethod,
		MatchPath:        assertion.matchPath,
		MatchHeaders:     assertion.matchHeaders,
		MatchParameters:  assertion.matchParameters,
		MatchBody:        assertion.matchBody,
		AtLeast:          assertion.atLeast,
		AtMost:           assertion.atMost,
		ReturnStatusCode: assertion.returnStatusCode,
		ReturnHeaders:    convertMapFromGolangToProto(assertion.returnHeaders),
		ReturnBody:       assertion.returnBody,
	}
}

func (assertion *Assertion) MatchHost(hostWithPort string) error {
	if assertion.matchHost {
		parts := strings.SplitN(hostWithPort, ":", 2)
		if assertion.host != parts[0] {
			return fmt.Errorf("Host mismatched: %s != %s", assertion.host, parts[0])
		}
	}
	return nil
}

func (assertion *Assertion) MatchMethod(method string) error {
	if assertion.matchMethod {
		expectedMethod := strings.ToUpper(assertion.method)
		actualMethod := strings.ToUpper(method)
		if expectedMethod != actualMethod {
			return fmt.Errorf("Method mismatched: %s != %s", expectedMethod, actualMethod)
		}
	}
	return nil
}

func (assertion *Assertion) MatchPath(path, rawPath string) error {
	if assertion.matchPath && assertion.path != path && assertion.path != rawPath {
		return fmt.Errorf("Path mismatched: %s != %s", assertion.path, path)
	}
	return nil
}

func (assertion *Assertion) MatchHeaders(headers http.Header) error {
	if assertion.matchHeaders {
		return matchKeyValues("Headers", map[string][]string(assertion.headers), map[string][]string(headers))
	}
	return nil
}

func (assertion *Assertion) MatchParameters(params url.Values) error {
	if assertion.matchParameters {
		return matchKeyValues("Parameters", map[string][]string(assertion.parameters), map[string][]string(params))
	}
	return nil
}

func matchKeyValues(name string, expected, actual map[string][]string) error {
	for expectedKey, expectedValues := range expected {
		if actualValues, exists := actual[expectedKey]; !exists {
			return fmt.Errorf("%s mismatched: %s is expected but not existed", expectedKey)
		} else {
		toNextValue:
			for _, expectedValue := range expectedValues {
				for _, actualValue := range actualValues {
					if expectedValue == actualValue {
						continue toNextValue
					}
				}
				return fmt.Errorf("%s mismatched: %s values does not match", expectedKey)
			}
		}
	}
	return nil
}

func (assertion *Assertion) MatchBody(body []byte) error {
	if assertion.matchBody && bytes.Compare(assertion.body, body) != 0 {
		fmt.Errorf("Body mismatched")
	}
	return nil
}

func (assertion *Assertion) MatchRequest(request *http.Request) error {
	if err := assertion.MatchHost(request.Host); err != nil {
		return err
	}
	if err := assertion.MatchMethod(request.Method); err != nil {
		return err
	}
	if err := assertion.MatchPath(request.URL.Path, request.URL.RawPath); err != nil {
		return err
	}
	if err := assertion.MatchHeaders(request.Header); err != nil {
		return err
	}
	if err := assertion.MatchParameters(request.URL.Query()); err != nil {
		return err
	}
	if requestBody, err := ioutil.ReadAll(request.Body); err != nil {
		return err
	} else if err = request.Body.Close(); err != nil {
		return err
	} else if err = assertion.MatchBody(requestBody); err != nil {
		return err
	}
	return nil
}
