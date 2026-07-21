package server

import (
	"testing"
)

func Test_linkify(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "plain text",
			text: "Hello, world!",
			want: "Hello, world!",
		},
		{
			name: "bare http url",
			text: "http://example.com",
			want: `<a href="http://example.com" target="_blank" rel="noopener noreferrer">http://example.com</a>`,
		},
		{
			name: "bare https url",
			text: "https://example.com",
			want: `<a href="https://example.com" target="_blank" rel="noopener noreferrer">https://example.com</a>`,
		},
		{
			name: "url surrounded by text",
			text: "Check this out: https://example.com/foo?bar=baz it's great",
			want: `Check this out: <a href="https://example.com/foo?bar=baz" target="_blank" rel="noopener noreferrer">https://example.com/foo?bar=baz</a> it&#39;s great`,
		},
		{
			name: "multiple urls",
			text: "https://foo.com and https://bar.com",
			want: `<a href="https://foo.com" target="_blank" rel="noopener noreferrer">https://foo.com</a> and <a href="https://bar.com" target="_blank" rel="noopener noreferrer">https://bar.com</a>`,
		},
		{
			name: "html is escaped, not rendered",
			text: `<script>alert("xss")</script>`,
			want: `&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`,
		},
		{
			name: "malicious href attribute is escaped, not injected",
			text: `" onmouseover="alert(1)`,
			want: `&#34; onmouseover=&#34;alert(1)`,
		},
		{
			name: "empty string",
			text: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(linkify(tt.text))
			if got != tt.want {
				t.Errorf("linkify(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}
