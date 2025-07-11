package functions

import (
	"context"
	"encoding/base64"
	"html"
	"testing"
)

func TestURLEncodeFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
	}{
		{
			name:      "simple string",
			args:      []interface{}{"hello world"},
			expected:  "hello+world",
			wantError: false,
		},
		{
			name:      "special characters",
			args:      []interface{}{"hello@world.com"},
			expected:  "hello%40world.com",
			wantError: false,
		},
		{
			name:      "empty string",
			args:      []interface{}{""},
			expected:  "",
			wantError: false,
		},
		{
			name:      "wrong number of arguments",
			args:      []interface{}{"arg1", "arg2", "arg3", "arg4"},
			expected:  "",
			wantError: true,
		},
		{
			name:      "non-string argument",
			args:      []interface{}{123},
			expected:  "",
			wantError: true,
		},
		{
			name:      "with custom unescaped characters",
			args:      []interface{}{"hello/world", "/"},
			expected:  "hello/world",
			wantError: false,
		},
		{
			name:      "with multiple custom unescaped characters",
			args:      []interface{}{"a=b&c=d", "&="},
			expected:  "a=b&c=d",
			wantError: false,
		},
		{
			name:      "double encode",
			args:      []interface{}{"hello world", 2.0},
			expected:  "hello%2Bworld",
			wantError: false,
		},
		{
			name:      "triple encode",
			args:      []interface{}{"hello world", 3.0},
			expected:  "hello%252Bworld",
			wantError: false,
		},
		{
			name:      "double encode with chars to not encode",
			args:      []interface{}{"hello/world", 2.0, "/"},
			expected:  "hello/world",
			wantError: false,
		},
		{
			name:      "invalid times argument type",
			args:      []interface{}{"hello", true},
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &URLEncodeFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestURLDecodeFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
	}{
		{
			name:      "encoded string",
			args:      []interface{}{"hello+world"},
			expected:  "hello world",
			wantError: false,
		},
		{
			name:      "encoded special characters",
			args:      []interface{}{"hello%40world.com"},
			expected:  "hello@world.com",
			wantError: false,
		},
		{
			name:      "invalid encoding",
			args:      []interface{}{"hello%ZZ"},
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &URLDecodeFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBase64EncodeFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
	}{
		{
			name:      "simple string",
			args:      []interface{}{"hello"},
			expected:  base64.StdEncoding.EncodeToString([]byte("hello")),
			wantError: false,
		},
		{
			name:      "empty string",
			args:      []interface{}{""},
			expected:  "",
			wantError: false,
		},
		{
			name:      "special characters",
			args:      []interface{}{"hello@world!"},
			expected:  base64.StdEncoding.EncodeToString([]byte("hello@world!")),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &Base64EncodeFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBase64DecodeFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
	}{
		{
			name:      "valid base64",
			args:      []interface{}{base64.StdEncoding.EncodeToString([]byte("hello"))},
			expected:  "hello",
			wantError: false,
		},
		{
			name:      "invalid base64",
			args:      []interface{}{"invalid!!!"},
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &Base64DecodeFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHTMLEncodeFunction(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected string
	}{
		{
			name:     "html special characters",
			args:     []interface{}{"<script>alert('xss')</script>"},
			expected: html.EscapeString("<script>alert('xss')</script>"),
		},
		{
			name:     "ampersand",
			args:     []interface{}{"cats & dogs"},
			expected: "cats &amp; dogs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &HTMLEncodeFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHTMLDecodeFunction(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected string
	}{
		{
			name:     "html entities",
			args:     []interface{}{"&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
			expected: html.UnescapeString("&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"),
		},
		{
			name:     "ampersand entity",
			args:     []interface{}{"cats &amp; dogs"},
			expected: "cats & dogs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &HTMLDecodeFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
