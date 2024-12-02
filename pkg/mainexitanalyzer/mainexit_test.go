package mainexitanalyzer

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func Test_Mainexitanalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "./...")
}
