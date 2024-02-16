CREATE TABLE IF NOT EXISTS Users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE,
    uname TEXT,
    phone TEXT,
    uaddress TEXT,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
)

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   IF row(NEW.*) IS DISTINCT FROM row(OLD.*) THEN
      NEW.updated_at = now(); 
      RETURN NEW;
   ELSE
      RETURN OLD;
   END IF;
END;
$$ language 'plpgsql';
-- 
CREATE TABLE IF NOT EXISTS UserNotes (
    id UUID PRIMARY KEY,
    title TEXT,
    content TEXT,
    user_id references Users(id),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
)

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   IF row(NEW.*) IS DISTINCT FROM row(OLD.*) THEN
      NEW.updated_at = now(); 
      RETURN NEW;
   ELSE
      RETURN OLD;
   END IF;
END;
$$ language 'plpgsql';
