package db

import (
	"dilogger/internal/push"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/pocketbase/pocketbase/tools/types"
)

func NewServer() *Server {
	godotenv.Load()
	app := pocketbase.New()
	notifier := push.NewNotificationApp()
	return &Server{app, nil, nil, nil, notifier, app.Logger()}
}

func NewCollection(name string, args ...any) *core.Collection {
	collection := core.NewBaseCollection(name)
	collection.ListRule = types.Pointer("")
	collection.ViewRule = types.Pointer("")

	switch name {
	case "products":
		collection.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
		})
		collection.Fields.Add(&core.NumberField{
			Name:     "stock",
			Required: true,
			OnlyInt:  true,
		})
	case "prices":
		productCollectionID := args[0].(string)
		collection.Fields.Add(&core.RelationField{
			Name:          "product",
			Required:      true,
			CascadeDelete: true,
			CollectionId:  productCollectionID,
		})
		collection.Fields.Add(&core.NumberField{
			Name:     "price",
			Required: true,
		})
	case "urls":
		accessRule := "@request.auth.id != ''"
		collection.CreateRule = types.Pointer(accessRule)
		collection.UpdateRule = types.Pointer(accessRule)
		collection.DeleteRule = types.Pointer(accessRule)
		collection.Fields.Add(&core.URLField{
			Name:     "url",
			Required: true,
		})
		collection.Fields.Add(&core.SelectField{
			Name:     "type",
			Required: true,
			Values:   []string{"wishlist", "product"},
		})
		collection.AddIndex("idx_"+security.RandomString(10), true, "url", "")
	}
	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	collection.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})
	return collection
}
