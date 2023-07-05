package internal

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type SuiteTester struct {
	// Include our basic suite logic.
	suite.Suite

	// Keep counts of how many times each method is run.
	SetupSuiteRunCount      int
	TearDownSuiteRunCount   int
	SetupTestRunCount       int
	TearDownTestRunCount    int
	TestOneRunCount         int
	TestTwoRunCount         int
	TestSubtestRunCount     int
	NonTestMethodRunCount   int
	SetupSubTestRunCount    int
	TearDownSubTestRunCount int

	SuiteNameBefore []string
	TestNameBefore  []string

	SuiteNameAfter []string
	TestNameAfter  []string

	TimeBefore []time.Time
	TimeAfter  []time.Time
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *SuiteTester) SetupSuite() {
	suite.SetupSuiteRunCount++
}

func (suite *SuiteTester) BeforeTest(suiteName, testName string) {
}

func (suite *SuiteTester) AfterTest(suiteName, testName string) {
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *SuiteTester) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

// The SetupTest method will be run before every test in the suite.
func (suite *SuiteTester) SetupTest() {
	suite.SetupTestRunCount++
}

// The TearDownTest method will be run after every test in the suite.
func (suite *SuiteTester) TearDownTest() {
	suite.TearDownTestRunCount++
}

// Every method in a testing suite that begins with "Test" will be run
// as a test.  TestOne is an example of a test.  For the purposes of
// this example, we've included assertions in the tests, since most
// tests will issue assertions.
func (suite *SuiteTester) TestOne() {
	beforeCount := suite.TestOneRunCount
	suite.TestOneRunCount++
	assert.Equal(suite.T(), suite.TestOneRunCount, beforeCount+1)
	suite.Equal(suite.TestOneRunCount, beforeCount+1)
}

// TestTwo is another example of a test.
func (suite *SuiteTester) TestTwo() {
	beforeCount := suite.TestTwoRunCount
	suite.TestTwoRunCount++
	assert.NotEqual(suite.T(), suite.TestTwoRunCount, beforeCount)
	suite.NotEqual(suite.TestTwoRunCount, beforeCount)
}

func (suite *SuiteTester) TestSkip() {
	suite.T().Skip()
}

// NonTestMethod does not begin with "Test", so it will not be run by
// testify as a test in the suite.  This is useful for creating helper
// methods for your tests.
func (suite *SuiteTester) NonTestMethod() {
	suite.NonTestMethodRunCount++
}

func (suite *SuiteTester) TestSubtest() {
	suite.TestSubtestRunCount++

	for _, t := range []struct {
		testName string
	}{
		{"first"},
		{"second"},
	} {
		suiteT := suite.T()
		suite.Run(t.testName, func() {
			// We should get a different *testing.T for subtests, so that
			// go test recognizes them as proper subtests for output formatting
			// and running individual subtests
			subTestT := suite.T()
			suite.NotEqual(subTestT, suiteT)
		})
		suite.Equal(suiteT, suite.T())
	}
}

func (suite *SuiteTester) TearDownSubTest() {
	suite.TearDownSubTestRunCount++
}

func (suite *SuiteTester) SetupSubTest() {
	suite.SetupSubTestRunCount++
}

func TestRunSuite(t *testing.T) {
	suiteTester := new(SuiteTester)
	suite.Run(t, suiteTester)

	// Normally, the test would end here.  The following are simply
	// some assertions to ensure that the Run function is working as
	// intended - they are not part of the example.

	// The suite was only run once, so the SetupSuite and TearDownSuite
	// methods should have each been run only once.
	assert.Equal(t, suiteTester.SetupSuiteRunCount, 1)
	assert.Equal(t, suiteTester.TearDownSuiteRunCount, 1)

	//assert.Equal(t, len(suiteTester.SuiteNameAfter), 4)
	//assert.Equal(t, len(suiteTester.SuiteNameBefore), 4)
	//assert.Equal(t, len(suiteTester.TestNameAfter), 4)
	//assert.Equal(t, len(suiteTester.TestNameBefore), 4)
	//
	//assert.Contains(t, suiteTester.TestNameAfter, "TestOne")
	//assert.Contains(t, suiteTester.TestNameAfter, "TestTwo")
	//assert.Contains(t, suiteTester.TestNameAfter, "TestSkip")
	//assert.Contains(t, suiteTester.TestNameAfter, "TestSubtest")
	//
	//assert.Contains(t, suiteTester.TestNameBefore, "TestOne")
	//assert.Contains(t, suiteTester.TestNameBefore, "TestTwo")
	//assert.Contains(t, suiteTester.TestNameBefore, "TestSkip")
	//assert.Contains(t, suiteTester.TestNameBefore, "TestSubtest")
	//
	//for _, suiteName := range suiteTester.SuiteNameAfter {
	//	assert.Equal(t, "SuiteTester", suiteName)
	//}
	//
	//for _, suiteName := range suiteTester.SuiteNameBefore {
	//	assert.Equal(t, "SuiteTester", suiteName)
	//}
	//
	//for _, when := range suiteTester.TimeAfter {
	//	assert.False(t, when.IsZero())
	//}
	//
	//for _, when := range suiteTester.TimeBefore {
	//	assert.False(t, when.IsZero())
	//}

	// There are four test methods (TestOne, TestTwo, TestSkip, and TestSubtest), so
	// the SetupTest and TearDownTest methods (which should be run once for
	// each test) should have been run four times.
	//assert.Equal(t, suiteTester.SetupTestRunCount, 4)
	//assert.Equal(t, suiteTester.TearDownTestRunCount, 4)
	//
	//// Each test should have been run once.
	//assert.Equal(t, suiteTester.TestOneRunCount, 1)
	//assert.Equal(t, suiteTester.TestTwoRunCount, 1)
	//assert.Equal(t, suiteTester.TestSubtestRunCount, 1)
	//
	assert.Equal(t, suiteTester.TearDownSubTestRunCount, 2)
	assert.Equal(t, suiteTester.SetupSubTestRunCount, 2)
	//
	//// Methods that don't match the test method identifier shouldn't
	//// have been run at all.
	//assert.Equal(t, suiteTester.NonTestMethodRunCount, 0)

}
