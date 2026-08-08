package hash

import "testing"

func TestSumSHA256(t *testing.T) {
	// Arrange
	data := []byte(`{"ok":true}`)
	key := "secret"
	want := "f6b4a2841c93f8bf2fb8f2c13d8fb0b6c8e8019f09ee405d248daa8385fad638"

	// Act
	got := SumSHA256(data, key)

	// Assert
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestVerifySHA256(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		key      string
		expected string
		want     bool
	}{
		{
			name:     "valid hash",
			data:     []byte(`{"ok":true}`),
			key:      "secret",
			expected: "f6b4a2841c93f8bf2fb8f2c13d8fb0b6c8e8019f09ee405d248daa8385fad638",
			want:     true,
		},
		{
			name:     "invalid hash",
			data:     []byte(`{"ok":true}`),
			key:      "secret",
			expected: "invalid",
			want:     false,
		},
		{
			name:     "wrong key",
			data:     []byte(`{"ok":true}`),
			key:      "wrong",
			expected: "f6b4a2841c93f8bf2fb8f2c13d8fb0b6c8e8019f09ee405d248daa8385fad638",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := VerifySHA256(tt.data, tt.key, tt.expected)

			// Assert
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
