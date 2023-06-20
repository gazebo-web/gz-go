package firestore

import (
	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestModel_FromDocumentSnapshot(t *testing.T) {
	var model Model
	now := time.Now()
	doc := &firestore.DocumentSnapshot{
		Ref: &firestore.DocumentRef{
			ID: "1234",
		},
		CreateTime: now,
		UpdateTime: now.Add(time.Hour),
		ReadTime:   now,
	}

	model = model.FromDocumentSnapshot(doc)
	assert.NotEmpty(t, model.ID)
	assert.False(t, model.CreatedAt.IsZero())
	assert.False(t, model.UpdatedAt.IsZero())

	assert.Equal(t, doc.Ref.ID, model.ID)
	assert.True(t, doc.CreateTime.Equal(model.CreatedAt))
	assert.True(t, doc.UpdateTime.Equal(model.UpdatedAt))
}
