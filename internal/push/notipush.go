package push

import (
	"context"
	"dilogger/internal/db"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/OneSignal/onesignal-go-api/v2"
	"github.com/google/uuid"
)

// The type Output has two fields, Id and External_id, with corresponding JSON tags.
type Output struct {
	Id          string `json:"id"`
	External_id string `json:"external_id"`
}

var (
	client  = onesignal.NewAPIClient(onesignal.NewConfiguration())
	authCtx = context.WithValue(context.Background(), onesignal.UserAuth, os.Getenv("OS_APP_KEY"))
	appId   = os.Getenv("OS_APP_ID")
)

// The `SendNotification` function sends a push notification using OneSignal with custom data and verifies the notification's external ID.
func SendNotification(data db.Product) {
	var input map[string]any
	noti := *onesignal.NewNotification(appId)
	eid := uuid.New().String()
	noti.SetExternalId(eid)
	noti.SetIsIos(false)
	noti.SetName("API Notification")
	noti.SetTemplateId("0bcfe003-96f2-4cff-9615-43ec4cd0d035")
	noti.SetIncludedSegments([]string{"Total Subscriptions"})
	_data, _ := json.Marshal(data)
	json.Unmarshal(_data, &input)
	noti.SetCustomData(input)

	_, r, err := client.DefaultApi.CreateNotification(authCtx).Notification(noti).Execute()

	if err != nil {
		log.Fatalln(err)
	}

	var out Output
	err = json.NewDecoder(r.Body).Decode(&out)

	if err != nil {
		log.Fatalln(err)
	}
	if out.External_id != eid {
		log.Fatalln("Invalid notification")
	}
	fmt.Printf("Pushed notification: %s\n", out.Id)
}
