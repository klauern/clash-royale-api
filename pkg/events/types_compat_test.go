package events

import (
	"encoding/json"
	"testing"
)

func TestBattleRecord_Unmarshal_BackwardCompatible(t *testing.T) {
	legacy := []byte(`{
		"timestamp":"2026-02-15T00:00:00Z",
		"opponent_tag":"#ABC123",
		"result":"win",
		"crowns":3,
		"opponent_crowns":1,
		"battle_mode":"Grand Challenge"
	}`)

	var record BattleRecord
	if err := json.Unmarshal(legacy, &record); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if record.OpponentTag != "#ABC123" {
		t.Errorf("OpponentTag = %q, want %q", record.OpponentTag, "#ABC123")
	}
	if record.PlayerDeckHash != "" {
		t.Errorf("PlayerDeckHash = %q, want empty for legacy payload", record.PlayerDeckHash)
	}
	if record.OpponentDeckHash != "" {
		t.Errorf("OpponentDeckHash = %q, want empty for legacy payload", record.OpponentDeckHash)
	}
}
