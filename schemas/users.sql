CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE,
    uname TEXT,
    phone TEXT,
    uaddress TEXT,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TRIGGER tr_users_bu BEFORE UPDATE on users
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();