package main

import (
	"github.com/gostaticanalysis/emptycase"
	"github.com/romanp1989/go-shortener/pkg/mainexitanalyzer"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1008"
	"honnef.co/go/tools/stylecheck/st1013"
)

func main() {
	myChecks := []*analysis.Analyzer{
		// defines an Analyzer that detects useless assignments
		assign.Analyzer,
		// defines an Analyzer that detects common mistakes involving boolean operators
		bools.Analyzer,
		// defers defines an Analyzer that checks for common mistakes in defer statements
		defers.Analyzer,
		// defines an Analyzer that checks for mistakes using HTTP responses
		httpresponse.Analyzer,
		// defines an Analyzer that checks consistency of Printf format strings and arguments
		printf.Analyzer,
		// defines an Analyzer that checks for shadowed variables
		shadow.Analyzer,
		// defines an Analyzer that checks struct field tags are well formed
		structtag.Analyzer,
		// defines an Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions
		unmarshal.Analyzer,
		// unreachable defines an Analyzer that checks for unreachable code
		unreachable.Analyzer,
		// defines an analyzer that checks for unused results of calls to certain pure functions
		unusedresult.Analyzer,
		// should use constants for HTTP error codes, not magic numbers
		st1013.Analyzer,
		// a function’s error value should be its last return value
		st1008.Analyzer,
		// checks whether HTTP response body is closed
		bodyclose.Analyzer,
		// finds case statements with no body
		emptycase.Analyzer,
		mainexitanalyzer.Analyzer,
	}

	for _, v := range staticcheck.Analyzers {
		myChecks = append(myChecks, v.Analyzer)
	}

	for _, v := range simple.Analyzers {
		myChecks = append(myChecks, v.Analyzer)
	}

	multichecker.Main(
		myChecks...,
	)
}
