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

type AsyncAssertion struct{}

func (a AsyncAssertion) Should(matcher interface{}, optionalDescription ...interface{}) bool {
	return false
}

func Succeed() bool {
	return false
}
