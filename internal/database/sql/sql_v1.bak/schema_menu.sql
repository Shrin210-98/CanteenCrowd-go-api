-- Solution 1
CREATE TABLE menus (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE, -- e.g., "Breakfast", "Summer Specials"
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    valid_from TIMESTAMP,
    valid_to TIMESTAMP
);

-- Hierarchical categories
CREATE TABLE menu_categories (
    id SERIAL PRIMARY KEY,
    menu_id INTEGER REFERENCES menus(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES menu_categories(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    display_order INTEGER NOT NULL DEFAULT 0
);

-- Menu items that can belong to multiple categories
CREATE TABLE menu_items (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    cost DECIMAL(10,2),
    preparation_time INTEGER,
    is_vegetarian BOOLEAN DEFAULT FALSE,
    is_vegan BOOLEAN DEFAULT FALSE,
    is_gluten_free BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    image_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Junction table for items to categories (many-to-many)
CREATE TABLE menu_item_categories (
    menu_item_id INTEGER REFERENCES menu_items(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES menu_categories(id) ON DELETE CASCADE,
    display_order INTEGER DEFAULT 0,
    PRIMARY KEY (menu_item_id, category_id)
);

-- Junction table for items to menus (if items can appear in multiple menus)
CREATE TABLE menu_menu_items (
    menu_id INTEGER REFERENCES menus(id) ON DELETE CASCADE,
    menu_item_id INTEGER REFERENCES menu_items(id) ON DELETE CASCADE,
    PRIMARY KEY (menu_id, menu_item_id)
);


-- 1. Indexing
-- Essential indexes for your relational schema
CREATE INDEX idx_menu_categories_menu_id ON menu_categories(menu_id);
CREATE INDEX idx_menu_categories_parent_id ON menu_categories(parent_id);
CREATE INDEX idx_menu_item_categories_item_id ON menu_item_categories(menu_item_id);
CREATE INDEX idx_menu_item_categories_category_id ON menu_item_categories(category_id);
CREATE INDEX idx_menu_items_active ON menu_items(is_active) WHERE is_active = TRUE;




-- Solution 2
-- Menu-specific categories (hierarchical within each menu)
CREATE TABLE menu_categories (
    id SERIAL PRIMARY KEY,
    menu_id INTEGER NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES menu_categories(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    display_order INTEGER NOT NULL DEFAULT 0,
    -- Ensure category names are unique within a menu
    UNIQUE(menu_id, name),
    -- Ensure parent category belongs to the same menu
    CONSTRAINT fk_menu_category_parent CHECK (
        parent_id IS NULL OR EXISTS (
            SELECT 1 FROM menu_categories parent 
            WHERE parent.id = parent_id AND parent.menu_id = menu_id
        )
    )
);

-- Junction table linking menu items to menu-specific categories
CREATE TABLE menu_item_categories (
    menu_id INTEGER NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    menu_item_id INTEGER NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES menu_categories(id) ON DELETE CASCADE,
    display_order INTEGER DEFAULT 0,
    PRIMARY KEY (menu_id, menu_item_id, category_id),
    -- Ensure category belongs to the specified menu
    CONSTRAINT fk_menu_category_match CHECK (
        EXISTS (
            SELECT 1 FROM menu_categories 
            WHERE id = category_id AND menu_id = menu_categories.menu_id
        )
    )
);

-- Indexes for performance
CREATE INDEX idx_menu_categories_menu_id ON menu_categories(menu_id);
CREATE INDEX idx_menu_categories_parent_id ON menu_categories(parent_id);
CREATE INDEX idx_menu_item_categories_composite ON menu_item_categories(menu_id, category_id, menu_item_id);
CREATE INDEX idx_menu_items_active ON menu_items(is_active) WHERE is_active = TRUE;



