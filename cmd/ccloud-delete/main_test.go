package main

import (
	"os"
	"reflect"
	"testing"
)

func openFile(filePath string) []byte {
	fileData, _ := os.ReadFile(filePath)
	return fileData
}

func TestTopicsFromResult(t *testing.T) {

	type test struct {
		input []byte
		want  []string
	}

	tests := []test{
		{
			input: openFile("../../example.result.json"),
			want: []string{
				"dev.topic.first",
				"dev.topic.second",
				"dev.topic.third",
				"dev.topic.fourth",
			},
		},
	}

	for _, tc := range tests {
		got := ParseTopicList(tc.input)
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("bad topic list from example file")
		}
	}
}
