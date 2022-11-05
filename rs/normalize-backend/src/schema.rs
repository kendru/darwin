table! {
    user (id) {
        id -> Uuid,
        email_address -> Text,
        password_digest -> Nullable<Bytea>,
        name -> Nullable<Text>,
        created_at -> Timestamp,
        updated_at -> Timestamp,
        email_verified_at -> Nullable<Timestamp>,
    }
}

table! {
    user_token (token) {
        user_id -> Uuid,
        token -> Text,
        created_at -> Timestamp,
        expires_at -> Timestamp,
        redeemed_at -> Nullable<Timestamp>,
        user_token_type_id -> Int2,
    }
}

table! {
    user_token_type (id) {
        name -> Nullable<Text>,
        id -> Int2,
    }
}

joinable!(user_token -> user (user_id));
joinable!(user_token -> user_token_type (user_token_type_id));

allow_tables_to_appear_in_same_query!(
    user,
    user_token,
    user_token_type,
);
