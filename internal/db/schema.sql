CREATE TABLE meows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name string NOT NULL,
  created_at timestamp NOT NULL DEFAULT NOW()
);
