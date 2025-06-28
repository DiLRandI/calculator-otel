CREATE TABLE calculator_history (
    id SERIAL PRIMARY KEY,
    input1 NUMERIC NOT NULL,
    input2 NUMERIC NOT NULL,
    result NUMERIC NOT NULL,
    operation VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
