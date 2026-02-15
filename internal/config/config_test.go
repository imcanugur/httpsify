package config

import (
	"reflect"
	"testing"
)

func TestParsePortRange(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    PortRange
		wantErr bool
	}{
		{
			name:    "single port",
			input:   "22",
			want:    PortRange{Start: 22, End: 22},
			wantErr: false,
		},
		{
			name:    "port range",
			input:   "135-139",
			want:    PortRange{Start: 135, End: 139},
			wantErr: false,
		},
		{
			name:    "port range with spaces",
			input:   " 135 - 139 ",
			want:    PortRange{Start: 135, End: 139},
			wantErr: false,
		},
		{
			name:    "port 1",
			input:   "1",
			want:    PortRange{Start: 1, End: 1},
			wantErr: false,
		},
		{
			name:    "port 65535",
			input:   "65535",
			want:    PortRange{Start: 65535, End: 65535},
			wantErr: false,
		},
		{
			name:    "invalid - zero",
			input:   "0",
			want:    PortRange{},
			wantErr: true,
		},
		{
			name:    "invalid - too high",
			input:   "65536",
			want:    PortRange{},
			wantErr: true,
		},
		{
			name:    "invalid - negative",
			input:   "-1",
			want:    PortRange{},
			wantErr: true,
		},
		{
			name:    "invalid - reversed range",
			input:   "139-135",
			want:    PortRange{},
			wantErr: true,
		},
		{
			name:    "invalid - not a number",
			input:   "abc",
			want:    PortRange{},
			wantErr: true,
		},
		{
			name:    "invalid - partial range",
			input:   "135-",
			want:    PortRange{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePortRange(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePortRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParsePortRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePortRanges(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []PortRange
		wantErr bool
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name:  "single port",
			input: "22",
			want:  []PortRange{{Start: 22, End: 22}},
		},
		{
			name:  "multiple ports",
			input: "22,25,445",
			want: []PortRange{
				{Start: 22, End: 22},
				{Start: 25, End: 25},
				{Start: 445, End: 445},
			},
		},
		{
			name:  "mixed ports and ranges",
			input: "22,25,135-139,445",
			want: []PortRange{
				{Start: 22, End: 22},
				{Start: 25, End: 25},
				{Start: 135, End: 139},
				{Start: 445, End: 445},
			},
		},
		{
			name:  "with spaces",
			input: " 22 , 25 , 135-139 , 445 ",
			want: []PortRange{
				{Start: 22, End: 22},
				{Start: 25, End: 25},
				{Start: 135, End: 139},
				{Start: 445, End: 445},
			},
		},
		{
			name:    "invalid port in list",
			input:   "22,abc,445",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePortRanges(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePortRanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePortRanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPortRangeContains(t *testing.T) {
	tests := []struct {
		name  string
		pr    PortRange
		port  int
		want  bool
	}{
		{
			name: "single port - match",
			pr:   PortRange{Start: 22, End: 22},
			port: 22,
			want: true,
		},
		{
			name: "single port - no match",
			pr:   PortRange{Start: 22, End: 22},
			port: 23,
			want: false,
		},
		{
			name: "range - start boundary",
			pr:   PortRange{Start: 135, End: 139},
			port: 135,
			want: true,
		},
		{
			name: "range - end boundary",
			pr:   PortRange{Start: 135, End: 139},
			port: 139,
			want: true,
		},
		{
			name: "range - middle",
			pr:   PortRange{Start: 135, End: 139},
			port: 137,
			want: true,
		},
		{
			name: "range - below",
			pr:   PortRange{Start: 135, End: 139},
			port: 134,
			want: false,
		},
		{
			name: "range - above",
			pr:   PortRange{Start: 135, End: 139},
			port: 140,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pr.Contains(tt.port); got != tt.want {
				t.Errorf("PortRange.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigIsPortAllowed(t *testing.T) {
	cfg := &Config{
		DenyPorts: []PortRange{
			{Start: 22, End: 22},
			{Start: 135, End: 139},
			{Start: 445, End: 445},
		},
		AllowRange: PortRange{Start: 1024, End: 65535},
	}

	tests := []struct {
		name string
		port int
		want bool
	}{
		{"allowed port 8000", 8000, true},
		{"allowed port 3000", 3000, true},
		{"allowed port 5173", 5173, true},
		{"allowed boundary 1024", 1024, true},
		{"allowed boundary 65535", 65535, true},
		{"denied port 22", 22, false},
		{"denied port 135", 135, false},
		{"denied port 137", 137, false},
		{"denied port 139", 139, false},
		{"denied port 445", 445, false},
		{"below allow range", 80, false},
		{"below allow range", 443, false},
		{"below allow range", 1023, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.IsPortAllowed(tt.port); got != tt.want {
				t.Errorf("Config.IsPortAllowed(%d) = %v, want %v", tt.port, got, tt.want)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		port    int
		wantErr bool
	}{
		{0, true},
		{1, false},
		{80, false},
		{443, false},
		{8000, false},
		{65535, false},
		{65536, true},
		{-1, true},
		{100000, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			err := ValidatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort(%d) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}
