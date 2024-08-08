DELETE FROM objects
WHERE "guild_id" = $1 AND "ticket_id" = $2;