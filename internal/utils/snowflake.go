package utils

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	nodes   sync.Map
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateID generates an 6-11 character short code using snowflake for uniqueness
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

	// Convert to base62 without truncation
	return toBase62(snowflakeID)
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

	// Ensure at least 6 characters by padding with leading 'a'
	for len(result) < 6 {
		result = append([]byte{charset[0]}, result...)
	}

	return string(result)
}
