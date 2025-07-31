-- schema.sql
-- Employees and related tables
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    manager_id INT NULL
);

CREATE TABLE positions (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    level INT,
    is_management BOOLEAN DEFAULT FALSE
);

CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone TEXT,
    address TEXT,
    hire_date DATE NOT NULL,
    termination_date DATE NULL,
    position_id INT NOT NULL REFERENCES positions(id),
    department_id INT NOT NULL REFERENCES departments(id),
    salary DECIMAL(10,2),
    emergency_contact_name TEXT,
    emergency_contact_phone TEXT,
    profile_description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- For PostgreSQL, we need a trigger to handle updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_employees_updated_at
BEFORE UPDATE ON employees
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

ALTER TABLE departments ADD CONSTRAINT fk_manager 
    FOREIGN KEY (manager_id) REFERENCES employees(id);

-- Inventory and suppliers
CREATE TABLE inventory_categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT
);

CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    contact_person TEXT,
    email TEXT,
    phone TEXT,
    address TEXT,
    preferred BOOLEAN DEFAULT FALSE,
    notes TEXT
);

CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    item_name TEXT NOT NULL,
    category_id INT NOT NULL REFERENCES inventory_categories(id),
    quantity DECIMAL(10,3) NOT NULL,
    unit TEXT NOT NULL,
    supplier_id INT REFERENCES suppliers(supplier_id),
    cost_per_unit DECIMAL(10,2) NOT NULL,
    par_level DECIMAL(10,3),
    reorder_level DECIMAL(10,3),
    last_restocked_date DATE,
    expiration_date DATE,
    storage_location TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_inventory_updated_at
BEFORE UPDATE ON inventory
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Menu and recipes
CREATE TABLE menu_categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    display_order INT
);

CREATE TABLE menu (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    category_id INT NOT NULL REFERENCES menu_categories(id),
    price DECIMAL(10,2) NOT NULL,
    cost DECIMAL(10,2),
    preparation_time INT,
    is_vegetarian BOOLEAN DEFAULT FALSE,
    is_vegan BOOLEAN DEFAULT FALSE,
    is_gluten_free BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    image_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_menu_updated_at
BEFORE UPDATE ON menu
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE recipe_ingredients (
    id SERIAL PRIMARY KEY,
    menu_item_id INT NOT NULL REFERENCES menu(id),
    inventory_id INT NOT NULL REFERENCES inventory(id),
    quantity DECIMAL(10,3) NOT NULL,
    unit TEXT NOT NULL,
    notes TEXT,
    UNIQUE (menu_item_id, inventory_id)
);

-- Additional tables
CREATE TABLE shifts (
    id SERIAL PRIMARY KEY,
    employee_id INT NOT NULL REFERENCES employees(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    role TEXT NOT NULL,
    notes TEXT
);

CREATE TABLE reservations (
    id SERIAL PRIMARY KEY,
    customer_name TEXT NOT NULL,
    customer_phone TEXT NOT NULL,
    customer_email TEXT,
    party_size INT NOT NULL,
    reservation_time TIMESTAMP NOT NULL,
    duration INT DEFAULT 90, -- minutes
    notes TEXT,
    status TEXT DEFAULT 'confirmed',
    table_assignment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inventory_transactions (
    id SERIAL PRIMARY KEY,
    inventory_id INT NOT NULL REFERENCES inventory(id),
    quantity_change DECIMAL(10,3) NOT NULL,
    transaction_type TEXT NOT NULL, -- 'purchase', 'usage', 'waste', 'adjustment'
    transaction_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_id INT REFERENCES employees(id),
    reference_id INT, -- could link to purchase order, menu item, etc.
    notes TEXT
);

-- Indexes for performance
CREATE INDEX idx_employee_department ON employees(department_id);
CREATE INDEX idx_employee_position ON employees(position_id);
CREATE INDEX idx_inventory_category ON inventory(category_id);
CREATE INDEX idx_menu_category ON menu(category_id);
CREATE INDEX idx_recipe_menu_item ON recipe_ingredients(menu_item_id);
CREATE INDEX idx_shift_employee ON shifts(employee_id);
CREATE INDEX idx_reservation_time ON reservations(reservation_time);