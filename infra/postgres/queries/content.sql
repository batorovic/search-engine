-- name: SearchContents :many
SELECT id, external_id, provider, title, type, published_at,
       views, likes, reactions, reading_time, score, tags, created_at, updated_at
FROM contents
WHERE (
        @query::text = '' OR 
        to_tsvector('english', title) @@ plainto_tsquery('english', @query::text)
    )
    AND (
        cardinality(@tags::text[]) = 0 OR 
        tags && @tags::text[]
    )
    AND (
        cardinality(@content_types::text[]) = 0 OR 
        type = ANY(@content_types::text[])
    )
ORDER BY
    CASE WHEN @sort_by::varchar = 'popularity' THEN (views + likes + reactions) END DESC,
    score DESC
LIMIT @page_limit::int OFFSET @page_offset::int;

-- name: CountSearchContents :one
SELECT COUNT(*)
FROM contents
WHERE (
        @query::text = '' OR 
        to_tsvector('english', title) @@ plainto_tsquery('english', @query::text)
    )
    AND (
        cardinality(@tags::text[]) = 0 OR 
        tags && @tags::text[]
    )
    AND (
        cardinality(@content_types::text[]) = 0 OR 
        type = ANY(@content_types::text[])
    );

-- name: UpsertContent :one
INSERT INTO contents (
    external_id, provider, title, type, published_at, raw_data,
    views, likes, reactions, reading_time, score, tags
) VALUES (
    @external_id, @provider, @title, @type, @published_at, @raw_data,
    @views, @likes, @reactions, @reading_time, @score, @tags
)
ON CONFLICT (provider, external_id)
DO UPDATE SET
    title = EXCLUDED.title,
    published_at = EXCLUDED.published_at,
    raw_data = EXCLUDED.raw_data,
    views = EXCLUDED.views,
    likes = EXCLUDED.likes,
    reactions = EXCLUDED.reactions,
    reading_time = EXCLUDED.reading_time,
    score = EXCLUDED.score,
    tags = EXCLUDED.tags,
    updated_at = NOW()
RETURNING id, created_at, updated_at;

-- name: GetContentByID :one
SELECT id, external_id, provider, title, type, published_at,
       views, likes, reactions, reading_time, score, created_at, updated_at
FROM contents
WHERE id = @id;

-- name: GetContentByExternalID :one
SELECT id, external_id, provider, title, type, published_at,
       views, likes, reactions, reading_time, score, created_at, updated_at
FROM contents
WHERE provider = @provider AND external_id = @external_id;

-- name: DeleteContent :exec
DELETE FROM contents WHERE id = @id;

-- name: SearchContentsByProvider :many
SELECT id, external_id, provider, title, type, published_at,
       views, likes, reactions, reading_time, score, tags, created_at, updated_at
FROM contents
WHERE provider = @provider::varchar
  AND (
    @query::text = '' OR
    to_tsvector('english', title) @@ plainto_tsquery('english', @query::text)
  )
ORDER BY score DESC
LIMIT @page_limit::int OFFSET @page_offset::int;
