package utils

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	nodes sync.Map
)

// GenerateID generates a new snowflake ID.
func GenerateID(machineID int64) string {
	if machineID < 0 || machineID > 1023 {
		return ""
	}
	// Check if the node already exists for the given machineID
	nodeValue, _ := nodes.LoadOrStore(machineID, func() *snowflake.Node {
		node, err := snowflake.NewNode(machineID)
		if err != nil {
			return nil
		}
		return node
	}())
	node := nodeValue.(*snowflake.Node)
	return node.Generate().Base64() // ConvertToBase62 converts the snowflake ID to a base62 string for use in URLs.
}
