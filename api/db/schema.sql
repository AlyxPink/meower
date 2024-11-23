CREATE TABLE authors (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  created_at timestamp NOT NULL DEFAULT NOW()
);
CREATE TABLE meows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  author_id UUID REFERENCES authors(id),
  content text NOT NULL,
  created_at timestamp NOT NULL DEFAULT NOW()
);
