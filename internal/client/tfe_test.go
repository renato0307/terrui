package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSearchText(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedTextSearch string
		expectedTagsSearch string
	}{
		{
			name:               "no tags",
			input:              "normalsearchstring",
			expectedTextSearch: "normalsearchstring",
			expectedTagsSearch: "",
		},
		{
			name:               "only tags",
			input:              "tags:12345",
			expectedTextSearch: "",
			expectedTagsSearch: "12345",
		},
		{
			name:               "tags and search string",
			input:              "normalsearchstring tags:12345",
			expectedTextSearch: "normalsearchstring",
			expectedTagsSearch: "12345",
		},
		{
			name:               "two tags and two search string",
			input:              "normalsearchstring tags:12345 tags:54321 normalsearchstring2",
			expectedTextSearch: "normalsearchstring normalsearchstring2",
			expectedTagsSearch: "54321", // only the last tag is returned
		},
		{
			name:               "list of tags",
			input:              "tags:12345,54321",
			expectedTextSearch: "",
			expectedTagsSearch: "12345,54321",
		},
		{
			name:               "alias t",
			input:              "t:12345",
			expectedTextSearch: "",
			expectedTagsSearch: "12345",
		},
		{
			name:               "alias tag",
			input:              "t:12345",
			expectedTextSearch: "",
			expectedTagsSearch: "12345",
		},
		{
			name:               "all empty",
			input:              "",
			expectedTextSearch: "",
			expectedTagsSearch: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			textSearch, tagsSearch := parseSearchText(tc.input)
			assert.Equal(t, tc.expectedTextSearch, textSearch)
			assert.Equal(t, tc.expectedTagsSearch, tagsSearch)
		})
	}

}
