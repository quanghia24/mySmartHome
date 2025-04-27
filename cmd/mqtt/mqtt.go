package mqtt

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/quanghia24/mySmartHome/types"
)

func NewClient() MQTT.Client {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("error loading .env file in mqtt")
	// }
	username := os.Getenv("AIOUSER")
	// MQTT broker URL for Adafruit IO
	broker := os.Getenv("BROKER")

	// MQTT client options
	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetUsername(username)
	opts.SetPassword("aio_somY90gmOI1pIeD8KpeTb1uLlSeE")
	opts.SetClientID("go-client-12345")

	opts.AutoReconnect = true
	opts.OnConnectionLost = func(client MQTT.Client, err error) {
		fmt.Println("Connection lost:", err)
	}
	// Message handler
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("Received message on topic %s: %s\n", msg.Topic(), msg.Payload())
		// You can add logic here to store in DB, trigger other services, etc.
	})

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Connection error:", token.Error())
		os.Exit(1)
	}

	fmt.Println("Connected to Adafruit IO")
	return client
}

func ResubscribeDevices(store types.DeviceStore, mqttClient MQTT.Client, logStore types.LogDeviceStore) error {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("error loading .env file in mqtt")
	// }
	username := os.Getenv("AIOUSER")

	devices, err := store.GetAllDevices()
	if err != nil {
		return err
	}

	for _, d := range devices {
		topic := fmt.Sprintf("%s/feeds/%s", username, d.FeedKey)
		fmt.Println("Subscribing to:", topic)

		token := mqttClient.Subscribe(topic, 0, func(client MQTT.Client, msg MQTT.Message) {
			fmt.Printf("Received message on %s: %s\n", msg.Topic(), msg.Payload())

			message := ""
			value := string(msg.Payload())
			switch d.Type {
			case "door":
				if value == "0" {
					message = fmt.Sprintf("[%s] got closed", d.Title)
				} else {
					message = fmt.Sprintf("[%s] got opened", d.Title)
				}
			case "fan":
				message = fmt.Sprintf("[%s]'s set at level: %s", d.Title, value)
			case "light":
				message = fmt.Sprintf("[%s]'s set color: %s", d.Title, value)
			}

			err = logStore.CreateLog(types.LogDevice{
				Type:     "onoff",
				Message:  message,
				DeviceID: d.FeedID,
				UserID:   d.UserID,
				Value:    value,
			})
			if err != nil {
				fmt.Printf("log creation err at mqtt:%v\n", err)
			}
		})

		if token.Wait() && token.Error() != nil {
			fmt.Println("Failed to subscribe:", token.Error())
		}
	}

	return nil
}

func ResubscribeSensors(store types.SensorStore, mqttClient MQTT.Client, planStore types.PlanStore, logStore types.LogSensorStore) error {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("error loading .env file in mqtt")
	// }
	username := os.Getenv("AIOUSER")

	sensors, err := store.GetAllSensor()
	if err != nil {
		return err
	}

	for _, d := range sensors {
		topic := fmt.Sprintf("%s/feeds/%s", username, d.FeedKey)
		fmt.Println("Subscribing to:", topic)

		token := mqttClient.Subscribe(topic, 0, func(client MQTT.Client, msg MQTT.Message) {
			fmt.Printf("Received message on %s: %s\n", msg.Topic(), msg.Payload())

			f, _ := strconv.ParseFloat(string(msg.Payload()), 32)

			// Round to 1 decimal place
			value := math.Round(f*10) / 10

			// check for plan -> threshold
			// fmt.Println("Check threshold for", d.FeedId, "with value of", value)
			plan, err := planStore.GetPlansByFeedID(d.FeedId)
			if err != nil {
				fmt.Println("Failed to get plans:", err)
			}
			if plan != nil {
				if plan.Lower != "" {
					lower, _ := strconv.ParseFloat(plan.Lower, 32)
					if lower > value {
						fmt.Println("WARNING!!! lower")
						err = logStore.CreateLogSensor(types.LogSensor{
							Type:     "warning",
							Message:  fmt.Sprintf("%f below the %f lower bound", value, lower),
							SensorID: d.FeedId,
							UserID:   d.UserID,
							Value:    string(msg.Payload()),
						})

						if err != nil {
							log.Println("sensor log create:", err)
						}
					}
				}
				if plan.Upper != "" {
					upper, _ := strconv.ParseFloat(plan.Upper, 32)
					if upper < value {
						fmt.Println("WARNING!!! upper")
						err = logStore.CreateLogSensor(types.LogSensor{
							Type:     "warning",
							Message:  fmt.Sprintf("%f exceed the %f upper bound", value, upper),
							SensorID: d.FeedId,
							UserID:   d.UserID,
							Value:    string(msg.Payload()),
						})

						if err != nil {
							log.Println("sensor log create:", err)
						}
					}
				}
			}

		})

		if token.Wait() && token.Error() != nil {
			fmt.Println("Failed to subscribe:", token.Error())
		}
	}

	return nil
}
