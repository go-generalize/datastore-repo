package a

import (
	"cloud.google.com/go/datastore"
)

type UserNoAlias struct {
	ID        *datastore.Key `json:"id,omitempty" datastore:"-" datastore_key:""`
	UserName  string         `json:"user_name,omitempty" datastore:"user_name"`
	CreatedAt int64          `json:"created_at,omitempty" datastore:"created_at"`
	UpdatedAt int64          `json:"updated_at,omitempty" datastore:"updated_at"`
}
