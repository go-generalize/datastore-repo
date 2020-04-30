package a

import (
	"cloud.google.com/go/datastore"
)

type UserNoAlias struct {
	ID        *datastore.Key `protobuf:"bytes,2,opt,name=users,proto3" json:"users,omitempty" datastore:"id" goon:"parent"` // want "[*]datastore.Key should have datastore:\"-\" tag"
	UserName  string         `json:"user_name,omitempty" datastore:"user_name"`
	CreatedAt int64          `json:"created_at,omitempty" datastore:"created_at"`
	UpdatedAt int64          `json:"updated_at,omitempty" datastore:"updated_at"`
}
