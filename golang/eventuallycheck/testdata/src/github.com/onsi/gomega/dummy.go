package gomega

// Dummy file for testing

func Eventually(actual interface{}, intervals ...interface{}) AsyncAssertion {
	return EventuallyWithOffset(0, actual, intervals...)
}

type AsyncAssertion struct{}

func (a AsyncAssertion) Should(matcher interface{}, optionalDescription ...interface{}) bool {
	return false
}

func EventuallyWithOffset(offset int, actual interface{}, intervals ...interface{}) AsyncAssertion {
	return AsyncAssertion{}
}

func Succeed() bool {
	return false
}
