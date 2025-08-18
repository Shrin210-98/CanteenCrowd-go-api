-- name: GetActiveMenus :many
SELECT * FROM menus 
WHERE is_active = TRUE 
AND (valid_from IS NULL OR valid_from <= NOW()) 
AND (valid_to IS NULL OR valid_to >= NOW());

-- name: GetMenuByID :one
SELECT * FROM menus WHERE id = $1;

-- name: CreateMenu :one
INSERT INTO menus (name, description, categories, valid_from, valid_to)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateMenu :one
UPDATE menus
SET 
    name = $2,
    description = $3,
    categories = $4,
    is_active = $5,
    valid_from = $6,
    valid_to = $7
WHERE id = $1
RETURNING *;

-- name: DeleteMenu :exec
DELETE FROM menus WHERE id = $1;



-- name: GetActiveMenuItems :many
SELECT * FROM menu_items WHERE is_active = TRUE;

-- name: GetMenuItemByID :one
SELECT * FROM menu_items WHERE id = $1;

-- name: GetMenuItemsByIDs :many
SELECT * FROM menu_items WHERE id = ANY($1::int[]);

-- name: CreateMenuItem :one
INSERT INTO menu_items (
    name, description, price, cost, preparation_time, 
    is_vegetarian, is_vegan, is_gluten_free, image_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateMenuItem :one
UPDATE menu_items
SET 
    name = $2,
    description = $3,
    price = $4,
    cost = $5,
    preparation_time = $6,
    is_vegetarian = $7,
    is_vegan = $8,
    is_gluten_free = $9,
    is_active = $10,
    image_url = $11,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteMenuItem :exec
DELETE FROM menu_items WHERE id = $1;

-- name: FilterMenuItems :many
SELECT * FROM menu_items
WHERE 
    is_active = TRUE
    AND (price <= $1 OR $1 IS NULL)
    AND (is_vegetarian = $2 OR $2 IS NULL)
    AND (is_vegan = $3 OR $3 IS NULL)
    AND (is_gluten_free = $4 OR $4 IS NULL)
    AND (preparation_time <= $5 OR $5 IS NULL)
ORDER BY 
    CASE WHEN $6 = 'price_asc' THEN price END ASC,
    CASE WHEN $6 = 'price_desc' THEN price END DESC,
    CASE WHEN $6 = 'prep_time_asc' THEN preparation_time END ASC,
    CASE WHEN $6 = 'prep_time_desc' THEN preparation_time END DESC,
    name ASC;




/* -- Extreme cases --
*/
-- name: GetMenuItemsByCategory :many
SELECT mi.* 
FROM menu_items mi
JOIN menus m ON mi.id = ANY(m.categories::jsonb->>$1::text::int[])
WHERE m.id = $2 AND mi.is_active = TRUE;

/* 
name: GetMenuWithItems :many
WITH menu_data AS (
    SELECT * FROM menus WHERE id = $1
),
menu_items_data AS (
    SELECT mi.* 
    FROM menu_items mi
    WHERE mi.id = ANY(
        SELECT jsonb_array_elements_text(categories)::int 
        FROM menu_data
    )
    AND mi.is_active = TRUE
)
SELECT 
    md.*,
    jsonb_agg(mi) AS items
FROM 
    menu_data md
LEFT JOIN 
    menu_items_data mi ON true
GROUP BY 
    md.id;
*/

-- name: AddItemToMenuCategory :exec
UPDATE menus
SET categories = jsonb_insert(
    categories, 
    ('{' || array_length(categories::jsonb->$2, 1) || '}')::text[], 
    $1::jsonb
)
WHERE id = $3;


