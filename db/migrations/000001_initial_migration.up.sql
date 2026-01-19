-- migration schema
CREATE TABLE IF NOT EXISTS schema_migrations
(
  version BIGINT NOT NULL
    CONSTRAINT schema_migrations_pkey
      PRIMARY KEY,
  dirty BOOLEAN NOT NULL
);

-- update trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
  RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
