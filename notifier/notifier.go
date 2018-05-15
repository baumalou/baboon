package notifier

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var notificationTimer *Notifier

type Notifier struct {
	ErrorTimer    int64
	DegradedTimer int64
}

func GetNotifier() *Notifier {
	if notificationTimer == nil {
		notificationTimer = &Notifier{ErrorTimer: 0, DegradedTimer: 0}
		return notificationTimer
	}
	return notificationTimer
}
func notificationNeedsToBeSent(notification string) bool {
	if strings.Contains(notification, "Degraded") {
		if time.Now().Unix()-notificationTimer.DegradedTimer > int64((30 * time.Minute).Seconds()) {
			notificationTimer.DegradedTimer = time.Now().Unix()
			return true
		}
		return false
	} else if strings.Contains(notification, "Error") {
		if time.Now().Unix()-notificationTimer.ErrorTimer > int64((30 * time.Minute).Seconds()) {
			notificationTimer.ErrorTimer = time.Now().Unix()
			return true
		}
		return false
	}
	return true

}
func (nt *Notifier) SendStatusNotification(vals, notification string) {
	if !notificationNeedsToBeSent(notification) {
		return
	}
	nt.SendNotification(vals + notification)
}

func (nt *Notifier) SendNotification(notification string) error {
	url := "https://chat.workshop21.ch/hooks/5zhbybp88jgwp88zanu9j4751w"
	fmt.Println("URL:>", url)

	var jsonStr = []byte(`
		{
			"text": "` + notification + `"
		}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	return nil
}
