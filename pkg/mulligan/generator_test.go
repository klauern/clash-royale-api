package mulligan

import "testing"

func TestSharedArchetypeSignals(t *testing.T) {
	t.Run("beatdown heavy win conditions", func(t *testing.T) {
		if !hasHeavyWinCondition([]string{"Golem"}) {
			t.Fatal("expected Golem to match beatdown heavy signal")
		}
		if hasHeavyWinCondition([]string{"Hog Rider"}) {
			t.Fatal("did not expect Hog Rider to match beatdown heavy signal")
		}
	})

	t.Run("siege buildings", func(t *testing.T) {
		if !hasSiegeBuilding([]string{"Xbow"}) {
			t.Fatal("expected Xbow to match siege signal")
		}
		if hasSiegeBuilding([]string{"Inferno Tower"}) {
			t.Fatal("did not expect Inferno Tower to match siege signal")
		}
	})

	t.Run("bridge spam win conditions", func(t *testing.T) {
		if !hasBridgeSpamWinCondition([]string{"Battle Ram"}) {
			t.Fatal("expected Battle Ram to match bridge-spam signal")
		}
		if !hasBridgeSpamWinCondition([]string{"Hog Rider"}) {
			t.Fatal("expected Hog Rider to match bridge-spam signal")
		}
	})
}
