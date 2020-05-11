package a

import (
	ds "cloud.google.com/go/datastore"
)

type UserMissingDatastoreTag struct {
	ID        *ds.Key `json:"id" datastore_key:""` // want "datastore key should have datastore:\"\" tag"
	UserName  string  `json:"user_name,omitempty" datastore:"user_name"`
	CreatedAt int64   `json:"created_at,omitempty" datastore:"created_at"`
	UpdatedAt int64   `json:"updated_at,omitempty" datastore:"updated_at"`
}
