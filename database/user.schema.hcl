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
  column "created_at" {
    null = false
    type = 
  }
  primary_key {
    columns = [column.id]
  }
  index "uni_users_email" {
    unique  = true
    columns = [column.email]
  }
}
schema "public" {
  comment = "standard public schema"
}
