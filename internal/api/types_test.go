package api

import (
	"encoding/json"
	"testing"
)

func TestBuildUploadStateUnmarshalString(t *testing.T) {
	var state BuildUploadState
	if err := json.Unmarshal([]byte(`"COMPLETE"`), &state); err != nil {
		t.Fatalf("unmarshal string state: %v", err)
	}
	if got := state.String(); got != "COMPLETE" {
		t.Fatalf("state = %q, want COMPLETE", got)
	}
}

func TestBuildUploadStateUnmarshalObject(t *testing.T) {
	var state BuildUploadState
	if err := json.Unmarshal([]byte(`{"state":"AWAITING_UPLOAD","errors":[],"warnings":[]}`), &state); err != nil {
		t.Fatalf("unmarshal object state: %v", err)
	}
	if got := state.String(); got != "AWAITING_UPLOAD" {
		t.Fatalf("state = %q, want AWAITING_UPLOAD", got)
	}
}

func TestBuildUploadStateUnmarshalObjectFallbacks(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{name: "code", json: `{"code":"COMPLETE"}`, want: "COMPLETE"},
		{name: "value", json: `{"value":"FAILED"}`, want: "FAILED"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var state BuildUploadState
			if err := json.Unmarshal([]byte(test.json), &state); err != nil {
				t.Fatalf("unmarshal object state: %v", err)
			}
			if got := state.String(); got != test.want {
				t.Fatalf("state = %q, want %s", got, test.want)
			}
		})
	}
}

func TestBuildUploadStateUnmarshalNull(t *testing.T) {
	state := BuildUploadState("COMPLETE")
	if err := json.Unmarshal([]byte(`null`), &state); err != nil {
		t.Fatalf("unmarshal null state: %v", err)
	}
	if got := state.String(); got != "" {
		t.Fatalf("state = %q, want empty", got)
	}
}
