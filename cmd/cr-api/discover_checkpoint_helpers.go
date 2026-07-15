package main

import (
	"errors"
	"fmt"

	"github.com/klauer/clash-royale-api/go/internal/playertag"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/urfave/cli/v3"
)

type discoverPlayerTag struct {
	input     string
	sanitized string
	canonical string
}

func discoverPlayerTagFromValue(playerTag string) (discoverPlayerTag, error) {
	sanitizedTag, err := playertag.Sanitize(playerTag)
	if err != nil {
		return discoverPlayerTag{}, err
	}

	return discoverPlayerTag{
		input:     playerTag,
		sanitized: sanitizedTag,
		canonical: "#" + sanitizedTag,
	}, nil
}

func discoverPlayerTagFromCommand(cmd *cli.Command) (discoverPlayerTag, error) {
	return discoverPlayerTagFromValue(cmd.String(discoverFlagTag))
}

type discoverCheckpointState struct {
	tag            discoverPlayerTag
	checkpointPath string
	checkpoint     deck.DiscoveryCheckpoint
}

func loadDiscoverCheckpointState(playerTag, missingMessagePrefix, missingMessageSuffix string) (discoverCheckpointState, error) {
	tag, err := discoverPlayerTagFromValue(playerTag)
	if err != nil {
		return discoverCheckpointState{}, err
	}

	checkpointPath := discoverCheckpointPath(tag.sanitized)
	checkpoint, err := deck.LoadDiscoveryCheckpoint(checkpointPath)
	if err != nil {
		if errors.Is(err, deck.ErrNoCheckpoint) {
			return discoverCheckpointState{}, errors.New(missingMessagePrefix + tag.sanitized + missingMessageSuffix)
		}
		if errors.Is(err, deck.ErrInvalidCheckpoint) {
			return discoverCheckpointState{}, fmt.Errorf("failed to parse checkpoint: %w", err)
		}
		return discoverCheckpointState{}, err
	}

	return discoverCheckpointState{
		tag:            tag,
		checkpointPath: checkpointPath,
		checkpoint:     checkpoint,
	}, nil
}

func loadDiscoverCheckpointStateFromCommand(
	cmd *cli.Command,
	missingMessagePrefix,
	missingMessageSuffix string,
) (discoverCheckpointState, error) {
	return loadDiscoverCheckpointState(cmd.String(discoverFlagTag), missingMessagePrefix, missingMessageSuffix)
}
