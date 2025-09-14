CREATE TABLE results (
    id SERIAL PRIMARY KEY,
    simulation_id VARCHAR(255) UNIQUE NOT NULL,
    task_id VARCHAR(512),
    org_name VARCHAR(512) NOT NULL,
    status VARCHAR(20) CHECK (status IN ('pending', 'processing', 'failed', 'success')) NOT NULL DEFAULT 'pending',
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_results_updated_at
    BEFORE UPDATE ON results
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX idx_simulation_id ON results(simulation_id);
CREATE INDEX idx_org_name ON results(org_name);
CREATE INDEX idx_status ON results(status);
CREATE INDEX idx_created_at ON results(created_at);
