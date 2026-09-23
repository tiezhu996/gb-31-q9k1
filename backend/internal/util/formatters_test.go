package util

import "testing"

func TestFormatStatusText(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"post pending", "pending", "待审核"},
		{"post approved", "approved", "已通过"},
		{"post rejected", "rejected", "已驳回"},
		{"meetup open", "open", "招募中"},
		{"meetup full", "full", "已满员"},
		{"unknown", "weird", "weird"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatStatusText(tt.input); got != tt.expect {
				t.Fatalf("FormatStatusText(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestFormatPostType(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"image", "图文"},
		{"video", "视频"},
		{"audio", "audio"},
	}
	for _, tt := range tests {
		if got := FormatPostType(tt.input); got != tt.expect {
			t.Fatalf("FormatPostType(%q) = %q, want %q", tt.input, got, tt.expect)
		}
	}
}

func TestFormatSpecies(t *testing.T) {
	if got := FormatSpecies("dog"); got != "狗狗" {
		t.Fatalf("FormatSpecies(dog) = %q", got)
	}
	if got := FormatSpecies("cat"); got != "猫咪" {
		t.Fatalf("FormatSpecies(cat) = %q", got)
	}
}
