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
