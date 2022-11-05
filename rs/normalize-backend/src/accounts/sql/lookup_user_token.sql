select
  u.*
from
  "user" u
  inner join user_token ut on u.id = ut.user_id
  inner join user_token_type utt on ut.user_token_type_id = utt.id
where
  utt."name" = $1 and
  ut.token = $2 and
  ut.redeemed_at is null and
  ut.expires_at > now()
limit 1;
