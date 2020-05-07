CREATE TABLE IF NOT EXISTS manifest_configurations
(
    id          bigint                   NOT NULL GENERATED BY DEFAULT AS IDENTITY,
    manifest_id bigint                   NOT NULL,
    media_type  text                     NOT NULL,
    digest_hex  bytea                    NOT NULL,
    size        bigint                   NOT NULL,
    payload     bytea                    NOT NULL,
    created_at  timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at  timestamp with time zone,
    CONSTRAINT pk_manifest_configurations PRIMARY KEY (id),
    CONSTRAINT fk_manifest_configurations_manifest_id FOREIGN KEY (manifest_id)
        REFERENCES manifests (id) ON DELETE CASCADE,
    CONSTRAINT uq_manifest_configurations_digest_hex UNIQUE (digest_hex),
    CONSTRAINT uq_manifest_configurations_manifest_id UNIQUE (manifest_id)
);