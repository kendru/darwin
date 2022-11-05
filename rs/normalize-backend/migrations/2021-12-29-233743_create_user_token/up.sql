create table user_token_type (
  "name" text,
  id smallint,
  primary key(id)
);

insert into user_token_type (id, "name") values
(1, 'login'),
(2, 'email');

-- token holds a temporary tokens that can be used for:
-- - email verification
-- - passwordless user login
create table user_token (
  user_id uuid not null,
  token text not null,
  created_at timestamp not null default now(),
  expires_at timestamp not null,
  redeemed_at timestamp null,
  user_token_type_id smallint not null,
  primary key (token),
  constraint fk_token_type foreign key (user_token_type_id) references "user_token_type" (id),
  constraint fk_user foreign key (user_id) references "user" (id) on delete cascade
);
