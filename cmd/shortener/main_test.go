package main

import "testing"

func Test_printBuildInfo(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "printBuildInfo_Success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printBuildInfo()
		})
	}
}
