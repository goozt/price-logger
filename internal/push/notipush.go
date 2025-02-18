package push

import (
	"context"
	"dilogger/internal/model"
	"dilogger/internal/utils"
	"encoding/json"
	"log"
	"strings"

	"github.com/OneSignal/onesignal-go-api/v2"
	"github.com/google/uuid"
)

// The type Output has two fields, Id and External_id, with corresponding JSON tags.
type Output struct {
	Id          string `json:"id"`
	External_id string `json:"external_id"`
}

type OneSignalApp struct {
	id       string
	client   *onesignal.APIClient
	authCtx  context.Context
	template string
	segments []string
}

func NewNotificationApp() *OneSignalApp {
	return &OneSignalApp{
		utils.GetEnv("OS_APP_ID"),
		onesignal.NewAPIClient(onesignal.NewConfiguration()),
		context.WithValue(context.Background(), onesignal.UserAuth, utils.GetEnv("OS_APP_KEY")),
		utils.GetEnv("OS_TEMPLATE_ID"),
		strings.Split(utils.GetEnv("OS_SEGMENT"), ","),
	}
}

// The `Send` function sends a push notification using OneSignal with custom data and verifies the notification's external ID.
func (app *OneSignalApp) Send(data model.Product) {
	var input map[string]any
	noti := *onesignal.NewNotification(app.id)
	eid := uuid.New().String()
	noti.SetExternalId(eid)
	noti.SetIsIos(false)
	noti.SetName("API Notification")
	noti.SetTemplateId(app.template)
	noti.SetIncludedSegments(app.segments)
	_data, _ := json.Marshal(data)
	json.Unmarshal(_data, &input)
	noti.SetCustomData(input)

	_, r, err := app.client.DefaultApi.CreateNotification(app.authCtx).Notification(noti).Execute()

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
	log.Printf("Pushed notification: %s\n", out.Id)
}
