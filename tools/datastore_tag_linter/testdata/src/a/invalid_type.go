package a

type UserInvalidType struct {
	ID        int    `json:"id,omitempty" datastore:"-" datastore_key:""` // want "available types for datastore key is [*]datastore.Key, string, int64"
	UserName  string `json:"user_name,omitempty" datastore:"user_name"`
	CreatedAt int64  `json:"created_at,omitempty" datastore:"created_at"`
	UpdatedAt int64  `json:"updated_at,omitempty" datastore:"updated_at"`
}
