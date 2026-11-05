package shard

import (
	"errors"
	"strconv"
)

const snowflakeShardShift = 22

func ForGuild(guildID string, shardCount int) (int, error) {
	if shardCount < 1 {
		return 0, errors.New("shard count must be positive")
	}
	id, err := strconv.ParseUint(guildID, 10, 64)
	if err != nil {
		return 0, errors.New("guild ID must be a numeric snowflake")
	}
	return int((id >> snowflakeShardShift) % uint64(shardCount)), nil
}
