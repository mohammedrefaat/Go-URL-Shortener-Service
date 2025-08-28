package utils

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

/*
var nodes sync.Map // map[int64]*snowflake.Node

// GenerateID generates a new snowflake ID in a concurrency-safe way.
func GenerateID(machineID int64) string {
	if machineID < 0 || machineID > 1023 {
		return ""
	}

	// Fast path: try load existing node
	if v, ok := nodes.Load(machineID); ok {
		if node, ok := v.(*snowflake.Node); ok && node != nil {
			return node.Generate().Base32()
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
			return existing.Generate().Base32()
		}
		// fallback to our created node if stored value is weird
	}
	// we stored 'node' successfully (or fallback)
	return node.Generate().Base32()
}
*/

var (
	nodes   sync.Map
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateID generates a 6-8 character short code using snowflake for uniqueness
func GenerateID(machineID int64) string {
	if machineID < 0 || machineID > 1023 {
		return ""
	}

	node := getOrCreateNode(machineID)
	if node == nil {
		return ""
	}

	// Generate snowflake ID
	snowflakeID := node.Generate().Int64()

	// Convert to base62 and truncate to 6-7 characters
	return toBase62(snowflakeID)[:7] // Take first 7 characters
}

func getOrCreateNode(machineID int64) *snowflake.Node {
	if v, ok := nodes.Load(machineID); ok {
		if node, ok := v.(*snowflake.Node); ok && node != nil {
			return node
		}
	}

	node, err := snowflake.NewNode(machineID)
	if err != nil {
		return nil
	}

	actual, _ := nodes.LoadOrStore(machineID, node)
	if existing, ok := actual.(*snowflake.Node); ok {
		return existing
	}
	return node
}

func toBase62(num int64) string {
	if num == 0 {
		return "aaaaaa" // 6-char default
	}

	base := int64(len(charset))
	var result []byte

	// Make positive for conversion
	if num < 0 {
		num = -num
	}

	for num > 0 {
		result = append([]byte{charset[num%base]}, result...)
		num /= base
	}

	// Ensure at least 6 characters
	for len(result) < 6 {
		result = append([]byte{charset[0]}, result...)
	}

	return string(result)
}
