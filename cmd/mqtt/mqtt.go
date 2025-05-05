package mqtt

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	"github.com/quanghia24/mySmartHome/services/device"
	"github.com/quanghia24/mySmartHome/services/log_device"
	"github.com/quanghia24/mySmartHome/services/log_sensor"
	"github.com/quanghia24/mySmartHome/services/notification"
	"github.com/quanghia24/mySmartHome/services/plan"
	"github.com/quanghia24/mySmartHome/services/sensor"
	"github.com/quanghia24/mySmartHome/types"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

func NewClient(db *sql.DB) MQTT.Client {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file in mqtt")
	}
	username := os.Getenv("AIOUSER")
	// MQTT broker URL for Adafruit IO
	broker := os.Getenv("BROKER")

	// MQTT client options
	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetUsername(username)
	opts.SetPassword(os.Getenv("AIOKey"))
	opts.SetClientID(os.Getenv("CLIENTID"))

	opts.AutoReconnect = true

	opts.OnConnect = func(client MQTT.Client) {
		fmt.Println("------- Trying to reconnecting to Adafruit IO -------")
		fmt.Println("connecting...")
		time.Sleep(2 * time.Second)

		deviceStore := device.NewStore(db)
		deviceLogStore := log_device.NewStore(db)

		sensorStore := sensor.NewStore(db)
		sensorLogStore := log_sensor.NewStore(db)
		planStore := plan.NewStore(db)
		notiStore := notification.NewStore(db)

		ResubscribeDevices(deviceStore, client, deviceLogStore)
		ResubscribeSensors(sensorStore, deviceStore, client, planStore, sensorLogStore, notiStore)
	}

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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file in mqtt")
	}
	username := os.Getenv("AIOUSER")

	devices, err := store.GetAllDevices()
	if err != nil {
		return err
	}

	for _, d := range devices {
		topic := fmt.Sprintf("%s/feeds/%s", username, d.FeedKey)
		// fmt.Println("Subscribing to:", topic)

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

	fmt.Println("done with device connections")
	return nil
}

func ResubscribeSensors(store types.SensorStore, deviceStore types.DeviceStore, mqttClient MQTT.Client, planStore types.PlanStore, logStore types.LogSensorStore, notiStore types.NotiStore) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file in mqtt")
	}
	username := os.Getenv("AIOUSER")

	sensors, err := store.GetAllSensor()
	if err != nil {
		return err
	}

	for _, d := range sensors {
		topic := fmt.Sprintf("%s/feeds/%s", username, d.FeedKey)
		// fmt.Println("Subscribing to:", topic)

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

						// check type
						mysensor, err := store.GetSensorByFeedID(d.FeedId)
						if err != nil {
							log.Println("error get sensor by id:", err)
						}

						if mysensor.Type == "brightness" {
							devices, err := deviceStore.GetDevicesInRoomID(d.RoomID)
							if err != nil {
								fmt.Println("error when get all devices in room:", err)
							}

							for _, device := range devices {
								if device.Type == "light" && device.Value == "#000000" {
									controlDevices(device)
								}
							}
						}

						// send out notification
						userIp, err := notiStore.GetNotiIpByUserId(d.UserID)
						if err != nil {
							fmt.Println(err)
						}
						if userIp != nil {
							msg := ""
							// send notification
							if mysensor.Type == "brightness" {
								msg = "trời tối thui rồi, tớ bật đèn nha :*"
							} else if mysensor.Type == "humidity" {
								msg = "khô quá, bộ nhà này cho lạc đà sống à :v"
							} else if mysensor.Type == "temperature" {
								msg = "lạnh quá người lạ ơi"
							}

							err := notiStore.CreateNoti(types.NotiPayload{
								UserID:  userIp.UserID,
								Ip:      userIp.Ip,
								Message: msg,
							})
							if err != nil {
								fmt.Println("error sending out notification")
							} else {
								sendNotification(userIp.Ip, msg)
							}
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

						// check type
						mysensor, err := store.GetSensorByFeedID(d.FeedId)
						if err != nil {
							log.Println("error get sensor by id:", err)
						}

						if mysensor.Type == "temperature" {
							devices, err := deviceStore.GetDevicesInRoomID(d.RoomID)
							if err != nil {
								fmt.Println("error when get all devices in room:", err)
							}

							for _, device := range devices {
								if device.Type == "fan" && device.Value == "0" {
									controlDevices(device)
								}
							}
						}

						// send out notification
						userIp, err := notiStore.GetNotiIpByUserId(d.UserID)
						if err != nil {
							fmt.Println(err)
						}
						if userIp != nil {
							msg := ""
							// send notification
							if mysensor.Type == "brightness" {
								msg = "Là Đảng hay sao mà chói thế"
							} else if mysensor.Type == "humidity" {
								msg = "Muốn trồng nấm hay gì?"
							} else if mysensor.Type == "temperature" {
								msg = "Con tép khô và cái lò nướng"
							}

							err := notiStore.CreateNoti(types.NotiPayload{
								UserID:  userIp.UserID,
								Ip:      userIp.Ip,
								Message: msg,
							})
							if err != nil {
								fmt.Println("error sending out notification")
							} else {
								sendNotification(userIp.Ip, msg)
							}
						}
					}
				}
			}

		})

		if token.Wait() && token.Error() != nil {
			fmt.Println("Failed to subscribe:", token.Error())
		}

	}
	fmt.Println("done with sensor connections")
	return nil
}

func controlDevices(device types.DeviceDataPayload) {
	url := os.Getenv("AIOAPI") + device.FeedKey + "/data"
	log.Println("adding data to", url)

	if device.Type == "fan" {
		device.Value = "75"
	} else if device.Type == "light" {
		device.Value = "#FFFFFF"
	} else {
		device.Value = "0"
	}

	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}

	device.CreatedAt = time.Now().In(location)

	// send request to adafruit server
	jsonData, err := json.Marshal(device)
	if err != nil {
		log.Fatalf("failed to marshal: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("failed to send request: %v", err)
		return
	}

	apiKey := os.Getenv("AIOKey")
	if apiKey == "" {
		fmt.Println("missing AIO Key")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIO-Key", apiKey)

	// make the request
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return
	}
}

func sendNotification(ip string, msg string) {
	pushToken, err := expo.NewExponentPushToken(fmt.Sprintf("ExponentPushToken[%s]", ip))
	if err != nil {
		panic(err)
	}

	// Create a new Expo SDK client
	client := expo.NewPushClient(nil)

	// Publish message
	response, err := client.Publish(
		&expo.PushMessage{
			To:   []expo.ExponentPushToken{pushToken},
			Body: msg,
			// Data: map[string]string{"withSome": "data"},
			Sound:    "default",
			Title:    "Warning",
			Priority: expo.DefaultPriority,
		},
	)

	// Check errors
	if err != nil {
		panic(err)
	}

	// Validate responses
	if response.ValidateResponse() != nil {
		fmt.Println(response.PushMessage.To, "failed")
	}
}
