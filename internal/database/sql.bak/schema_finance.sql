CREATE TABLE chart_of_accounts (
    id SERIAL PRIMARY KEY,
    account_code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('Asset', 'Liability', 'Equity', 'Revenue', 'Expense'))
);

CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_date DATE NOT NULL DEFAULT CURRENT_DATE,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE journal_lines (
    id SERIAL PRIMARY KEY,
    journal_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id INT NOT NULL REFERENCES chart_of_accounts(id),
    debit DECIMAL(12,2) DEFAULT 0,
    credit DECIMAL(12,2) DEFAULT 0,
    CHECK (debit >= 0 AND credit >= 0)
);
