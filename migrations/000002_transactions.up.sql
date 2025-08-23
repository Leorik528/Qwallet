


CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_from_address FOREIGN KEY (from_address) REFERENCES qwallets(address),
    CONSTRAINT fk_to_address FOREIGN KEY (to_address) REFERENCES qwallets(address)
);


-- Индекс для быстрого поиска транзакций по адресу
CREATE INDEX idx_transactions_from_address ON transactions(from_address);
CREATE INDEX idx_transactions_to_address ON transactions(to_address);