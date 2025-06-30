package schemabuilder

import (
	"testing"
	"time"

	"github.com/Rick-Phoenix/gofirst/db/sqlgen"
	"github.com/stretchr/testify/assert"
)

type UserWithPostsConst struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `dbignore:"true" json:"created_at"`
	Posts     []sqlgen.Post `json:"posts"`
}

var ValidUserSchema = MessageSchema{
	Name: "User",
	Fields: FieldsMap{
		1: String("name").Required().MinLen(2).MaxLen(32),
		2: Int64("id"),
		3: Timestamp("created_at"),
		4: Repeated("posts", MsgField("post", PostSchema)),
	},
	Model:      &UserWithPostsConst{},
	ImportPath: "myapp/v1/user.proto",
}

func TestModelValidation(t *testing.T) {
	err := ValidUserSchema.CheckModel()

	assert.NoError(t, err, "Testing model validation")
}
