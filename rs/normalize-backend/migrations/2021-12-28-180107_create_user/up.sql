create table "user" (
  id uuid primary key not null default gen_random_uuid(),
  email_address text not null,
  password_digest bytea,
  "name" text,
  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  email_verified_at timestamp null,
  unique(email_address)
);
select diesel_manage_updated_at('user');
