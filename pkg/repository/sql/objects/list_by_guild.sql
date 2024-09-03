SELECT "guild_id", "ticket_id", "bucket_id"
FROM objects
WHERE "guild_id" = $1;