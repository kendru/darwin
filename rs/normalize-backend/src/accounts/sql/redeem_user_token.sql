update user_token ut
set redeemed_at = now()
from
  "user" u,
  user_token_type utt
where
  u.id = ut.user_id and
  ut.user_token_type_id = utt.id and
  ut.expires_at > now() and
  ut.redeemed_at is null and
  utt."name" = $1 and
  ut.token = $2
returning u.*;
