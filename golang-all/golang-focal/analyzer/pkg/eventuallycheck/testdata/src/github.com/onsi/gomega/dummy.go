package gomega

// Dummy file for testing

func Consistently(actual interface{}, intervals ...interface{}) AsyncAssertion {
	return AsyncAssertion{}
}

func ConsistentlyWithOffset(offset int, actual interface{}, intervals ...interface{}) AsyncAssertion {
	return AsyncAssertion{}
}

func Eventually(actual interface{}, intervals ...interface{}) AsyncAssertion {
	return AsyncAssertion{}
}

func EventuallyWithOffset(offset int, actual interface{}, intervals ...interface{}) AsyncAssertion {
	return AsyncAssertion{}
}

func Expect(actual interface{}, extra ...interface{}) Assertion {
	return Assertion{}
}

func ExpectWithOffset(offset int, actual interface{}, extra ...interface{}) Assertion {
	return Assertion{}
}

func Ω(actual interface{}, extra ...interface{}) Assertion {
	return Assertion{}
}

type Assertion struct{}

type AsyncAssertion struct{}

func (a AsyncAssertion) Should(matcher interface{}, optionalDescription ...interface{}) bool {
	return false
}

func (a Assertion) To(matcher interface{}, optionalDescription ...interface{}) bool {
	return false
}

func BeTrue() bool {
	return false
}

func Succeed() bool {
	return false
}
