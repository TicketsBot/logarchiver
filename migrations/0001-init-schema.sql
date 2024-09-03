CREATE TABLE buckets
(
    "id"           uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    "endpoint_url" VARCHAR(255) NOT NULL,
    "name"         VARCHAR(255) NOT NULL,
    "active"       BOOLEAN      NOT NULL DEFAULT FALSE
);

CREATE TABLE objects
(
    "guild_id"  int8 NOT NULL,
    "ticket_id" int4 NOT NULL,
    "bucket_id" uuid NOT NULL,
    PRIMARY KEY ("guild_id", "ticket_id"),
    FOREIGN KEY ("bucket_id") REFERENCES "buckets" ("id")
);

CREATE INDEX objects_guid_id_idx ON objects ("guild_id");
