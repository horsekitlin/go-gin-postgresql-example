table "users" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "username" {
    null = true
    type = text
  }
  column "email" {
    null = true
    type = text
  }
  primary_key {
    columns = [column.id]
  }
  index "uni_users_username" {
    unique  = true
    columns = [column.username]
  }
}
schema "public" {
  comment = "standard public schema"
}
