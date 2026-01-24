package analyscheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestCheckAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, CheckAnalyzer, "a", "b", "c")
}
