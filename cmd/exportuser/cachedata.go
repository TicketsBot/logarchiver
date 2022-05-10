package main

import (
	"context"
	"encoding/json"
	"github.com/rxdn/gdl/cache"
	"strconv"
)

func getCacheData(cache *cache.PgCache, userId uint64) map[string]interface{} {
	data := make(map[string]interface{})

	user, ok := cache.GetUser(userId)
	if ok {
		data["user"] = user
	} else {
		data["user"] = nil
	}

	rows, err := cache.Query(context.Background(), `SELECT guild_id, data FROM members WHERE "user_id" = $1;`, userId)
	must(err)

	memberData := make(map[string]interface{})
	for rows.Next() {
		var guildId uint64
		var raw string

		must(rows.Scan(&guildId, &raw))

		memberData[strconv.FormatUint(guildId, 10)] = json.RawMessage([]byte(raw))
	}

	data["member_data"] = memberData

	return data
}
