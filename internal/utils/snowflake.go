package utils

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

var nodes sync.Map // map[int64]*snowflake.Node

// GenerateID generates a new snowflake ID in a concurrency-safe way.
func GenerateID(machineID int64) string {
	if machineID < 0 || machineID > 1023 {
		return ""
	}

	// Fast path: try load existing node
	if v, ok := nodes.Load(machineID); ok {
		if node, ok := v.(*snowflake.Node); ok && node != nil {
			return node.Generate().Base64()
		}
	}

	// Slow path: create a new node, then attempt to store it.
	// If another goroutine stored one first, use the stored one instead.
	node, err := snowflake.NewNode(machineID)
	if err != nil {
		return ""
	}

	actual, loaded := nodes.LoadOrStore(machineID, node)
	if loaded {
		// someone else stored a node first — use that one
		if existing, ok := actual.(*snowflake.Node); ok && existing != nil {
			return existing.Generate().Base64()
		}
		// fallback to our created node if stored value is weird
	}
	// we stored 'node' successfully (or fallback)
	return node.Generate().Base64()
}
