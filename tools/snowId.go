package tools

import (
	"github.com/bwmarrin/snowflake"
)

func GetSnowflakeId() int64 {
	//default node id eq 1,this can modify to different serverId node
	node, _ := snowflake.NewNode(1)
	// Generate a snowflake ID.
	id := node.Generate().Int64()
	return id
}
