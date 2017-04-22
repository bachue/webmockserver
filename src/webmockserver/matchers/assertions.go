package matchers

import (
	proto "webmockserver/proto"
)

type Assertions struct {
	assertions []*Assertion
}

func NewAssertionsFromProtobuf(assertions *proto.WebMockAssertions) *Assertions {
	retval := new(Assertions)
	retval.assertions = make([]*Assertion, 0, len(assertions.Assertions))

	for _, assertion := range assertions.Assertions {
		retval.assertions = append(retval.assertions, NewAssertionFromProtobuf(assertion))
	}

	return retval
}

func (assertions *Assertions) ToProtoBuffer() *proto.WebMockAssertions {
	retval := new(proto.WebMockAssertions)
	retval.Assertions = make([]*proto.WebMockAssertion, 0, len(assertions.assertions))

	for _, assertion := range assertions.assertions {
		retval.Assertions = append(retval.Assertions, assertion.ToProtoBuffer())
	}

	return retval
}
