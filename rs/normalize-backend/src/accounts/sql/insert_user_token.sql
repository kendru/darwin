insert into user_token (
  user_id,
  token,
  user_token_type_id,
  expires_at
) values (
  (select id from "user" where email_address = $1 limit 1),
  $2,
  (select id from user_token_type where "name" = $3 limit 1),
  now() + $4
);
