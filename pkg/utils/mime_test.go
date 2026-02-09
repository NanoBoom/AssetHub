package utils

import "testing"

func TestGetExtensionFromMIME(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     string
	}{
		{
			name:     "image/png",
			mimeType: "image/png",
			want:     ".png",
		},
		{
			name:     "image/jpeg",
			mimeType: "image/jpeg",
			want:     ".jpg",
		},
		{
			name:     "video/mp4",
			mimeType: "video/mp4",
			want:     ".mp4",
		},
		{
			name:     "application/pdf",
			mimeType: "application/pdf",
			want:     ".pdf",
		},
		{
			name:     "text/plain",
			mimeType: "text/plain",
			want:     ".txt",
		},
		{
			name:     "unknown type",
			mimeType: "application/x-unknown",
			want:     ".bin",
		},
		{
			name:     "with charset",
			mimeType: "text/plain; charset=utf-8",
			want:     ".txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExtensionFromMIME(tt.mimeType)
			if got != tt.want {
				t.Errorf("GetExtensionFromMIME(%q) = %q, want %q", tt.mimeType, got, tt.want)
			}
		})
	}
}
