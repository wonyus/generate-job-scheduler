package initials

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/wonyus/generate-job-scheduler/utils"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}

func InitialMqttClient(rev mqtt.MessageHandler) mqtt.Client {
	recover()
	var broker = utils.Strip(os.Getenv("MQTT_BROKER_URL"))
	portStr := utils.Strip(os.Getenv("MQTT_BROKER_PORT"))
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(utils.Strip(os.Getenv("MQTT_CLIENT_ID")))
	opts.SetUsername(utils.Strip(os.Getenv("MQTT_USERNAME")))
	opts.SetPassword(utils.Strip(os.Getenv("MQTT_PASSWORD")))
	opts.SetDefaultPublishHandler(rev)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetKeepAlive(time.Second * 60)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		panic(token.Error())
	}
	return client
}
