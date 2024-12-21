package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertBodyToContent(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []*Content
	}{
		{
			name: "Test 1",
			body: "Hello, **world**!",
			want: []*Content{
				{
					Text: "Hello, ",
				},
				{
					Tag: "i",
					Text: "world",
				},
				{
					Text: "!",
				},
			},
		},
		{
			name: "Test 2",
			body: "Hello, __world__!",
			want: []*Content{
				{
					Text: "Hello, ",
				},
				{
					Tag: "a",
					Text: "world",
				},
				{
					Text: "!",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertBodyToContent(tt.body)
			assert.Equal(t, tt.want, got)
		})
	}
} 