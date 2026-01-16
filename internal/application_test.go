package internal

import "testing"

func TestCreateAddr(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{name: "simple", host: "localhost", port: 8080, want: "localhost:8080"},
		{name: "ip", host: "127.0.0.1", port: 80, want: "127.0.0.1:80"},
		{name: "ipv6", host: "[::1]", port: 443, want: "[::1]:443"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateAddr(tt.host, tt.port)
			if got != tt.want {
				t.Fatalf("CreateAddr(%s, %d) = %s, want %s", tt.host, tt.port, got, tt.want)
			}
		})
	}
}
