-- 001_create_actual_layers_by_client_id.up.sql
CREATE TABLE IF NOT EXISTS actual_layers_by_client_id (
    clientID UUID,
    createdAt DateTime,
    updatedAt DateTime,
    layer1 Array(Float64),
    layer2 Array(Float64),
    layer3 Array(Float64),
    layer4 Array(Float64),
    layer5 Array(Float64)
    ) ENGINE = MergeTree()
ORDER BY (clientID, createdAt);
