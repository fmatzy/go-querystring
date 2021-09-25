package querystring

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func TestParse(t *testing.T) {
	tests := []struct {
		query   string
		t       transform.Transformer
		want    url.Values
		wantErr bool
	}{
		{
			query: "a=1&b=2",
			t:     japanese.ShiftJIS.NewDecoder(),
			want:  url.Values{"a": []string{"1"}, "b": []string{"2"}},
		},
		{
			query: "a=1&a=2&a=banana",
			t:     japanese.ShiftJIS.NewDecoder(),
			want:  url.Values{"a": []string{"1", "2", "banana"}},
		},
		{
			query: "ascii=%3Ckey%3A+0x90%3E",
			t:     japanese.ShiftJIS.NewDecoder(),
			want:  url.Values{"ascii": []string{"<key: 0x90>"}},
		},
		{
			query: "sjis=%3C%93%FA%96%7B%8C%EA+SJIS%3E",
			t:     japanese.ShiftJIS.NewDecoder(),
			want:  url.Values{"sjis": []string{"<日本語 SJIS>"}},
		},
		{
			query: "utf8=%3C%E6%97%A5%E6%9C%AC%E8%AA%9E+UTF8%3E",
			t:     unicode.UTF8.NewDecoder(),
			want:  url.Values{"utf8": []string{"<日本語 UTF8>"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := Parse(tt.query, tt.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		query   url.Values
		t       transform.Transformer
		want    string
		wantErr bool
	}{
		{
			query: url.Values{"sjis": []string{"<日本語 SJIS>"}},
			t:     japanese.ShiftJIS.NewEncoder(),
			want:  "sjis=%3C%93%FA%96%7B%8C%EA+SJIS%3E",
		},
		{
			query: url.Values{"utf8": []string{"<日本語 UTF8>"}},
			t:     unicode.UTF8.NewDecoder(),
			want:  "utf8=%3C%E6%97%A5%E6%9C%AC%E8%AA%9E+UTF8%3E",
		},
		{
			query:   url.Values{"sjis": []string{"<変換不可 〜>"}},
			t:       japanese.ShiftJIS.NewEncoder(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got, err := Encode(tt.query, tt.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Encode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
