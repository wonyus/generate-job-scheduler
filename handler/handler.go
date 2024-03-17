package handler

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/wonyus/generate-job-scheduler/context"
	"github.com/wonyus/generate-job-scheduler/repository"
)

type Handler struct {
	Context     context.Context
	Repository1 repository.Repository
	Repository2 repository.Repository
}

func (h Handler) Handler(batchSize int) error {
	// Array of JSON data
	var err error

	// Get switch config
	switches, err := h.Repository1.GetSwitchConfig()
	if err != nil {
		return fmt.Errorf("error getting switch config: %v", err)
	}

	// JSON data array
	var jsonDataArray [][]byte
	for _, val := range switches {
		jsonData := []byte(val.Scheduler.([]uint8))
		jsonDataArray = append(jsonDataArray, jsonData)
	}

	// Define a wait group to wait for goroutines to finish
	var wg sync.WaitGroup

	// Results slice to store messages
	var results []string
	today := time.Now()
	// Process JSON objects in batches
	for i := 0; i < len(jsonDataArray); i += batchSize {
		// Increment the wait group counter
		for j := i; j < i+batchSize && j < len(jsonDataArray); j++ {
			wg.Add(1)
		}

		// Process JSON objects concurrently
		for j := i; j < i+batchSize && j < len(jsonDataArray); j++ {
			go h.processScheduler(switches[j], jsonDataArray[j], &results, today, &wg)
		}

		// Wait for the goroutines to finish processing
		wg.Wait()
	}

	// Log report
	fmt.Printf("Today: %v\n", today.Format("2006-01-02 15:04:05"))
	fmt.Println("Amout of schedule Today: ", len(results))

	return nil
}

func (h Handler) processScheduler(data interface{}, jsonData []byte, results *[]string, today time.Time, wg *sync.WaitGroup) {
	defer wg.Done()

	// Unmarshal JSON data into the struct
	var scheduler repository.SchedulerValues
	if err := json.Unmarshal(jsonData, &scheduler); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	// Check if today is true in Days, Dates, and Months
	for _, day := range scheduler.Days {
		if day[0].(float64) == float64(today.Weekday()) && day[1].(bool) {
			for _, month := range scheduler.Months {
				if month[0].(float64) == float64(today.Month()) && month[1].(bool) {
					for _, date := range scheduler.Dates {
						if date[0].(float64) == float64(today.Day()) && date[1].(bool) {
							h.setScheduler(data, scheduler, today)
							*results = append(*results, "Scheduler set for today")
							return
						}
					}
					return
				}
			}
			return
		}
	}

}

func (h Handler) setScheduler(data interface{}, scheduler repository.SchedulerValues, today time.Time) error {
	newData := data.(repository.Switch)

	// Get scheduled time
	for _, value := range scheduler.Times {
		// Get scheduled time on
		t1, err := time.Parse(time.RFC3339, value[0])

		if err != nil {
			return fmt.Errorf("error parsing time: %v", err)
		}
		scheduledTime := time.Date(today.Year(), today.Month(), today.Day(), t1.Hour(), t1.Minute(), t1.Second(), 0, t1.Location())
		// scheduledTime, err := time.Parse(time.RFC3339, today.Format(time.RFC3339))
		// if err != nil {
		// 	return fmt.Errorf("error parsing time: %v", err)
		// }

		// Set scheduled time
		unixTimestamp := time.Unix(scheduledTime.Unix(), 0)
		commandObj, err := mapSwitchAction(newData, true)
		if err != nil {
			return fmt.Errorf("error mapping switch action: %v", err)
		}
		jsonBytes, err := json.Marshal(commandObj)
		if err != nil {
			return fmt.Errorf("error marshaling commandStr: %v", err)
		}
		jsonStr := string(jsonBytes)
		commandStr := fmt.Sprintf("%s|/%s/%s/%s|[%s]", "mqtt", "command", newData.MqttClientID, "switch", jsonStr)
		var task = repository.Task{Command: commandStr, ScheduledAt: pgtype.Timestamp{Time: convertDateToUTC(unixTimestamp)}}

		// Insert task
		err = h.insertTask(task)

		if err != nil {
			return fmt.Errorf("error inserting task: %v", err)
		}

		// Get scheduled time Off
		t1, err = time.Parse(time.RFC3339, value[1])
		if err != nil {
			return fmt.Errorf("error parsing time: %v", err)
		}
		scheduledTime = time.Date(today.Year(), today.Month(), today.Day(), t1.Hour(), t1.Minute(), t1.Second(), 0, t1.Location())
		// if err != nil {
		// 	return fmt.Errorf("error parsing time: %v", err)
		// }

		// Set scheduled time
		unixTimestamp = time.Unix(scheduledTime.Unix(), 0)
		commandObj, err = mapSwitchAction(newData, false)
		if err != nil {
			return fmt.Errorf("error mapping switch action: %v", err)
		}
		jsonBytes, err = json.Marshal(commandObj)
		if err != nil {
			return fmt.Errorf("error marshaling commandStr: %v", err)
		}
		jsonStr = string(jsonBytes)
		commandStr = fmt.Sprintf("%s|/%s/%s/%s|[%s]", "mqtt", "command", newData.MqttClientID, "switch", jsonStr)
		task = repository.Task{Command: commandStr, ScheduledAt: pgtype.Timestamp{Time: convertDateToUTC(unixTimestamp)}}

		// Insert task
		err = h.insertTask(task)

		if err != nil {
			return fmt.Errorf("error inserting task: %v", err)
		}
	}

	return nil
}

func (h Handler) insertTask(task repository.Task) error {
	// Insert task
	_, err := h.Repository2.InsertTask(task)

	if err != nil {
		return fmt.Errorf("error inserting task: %v", err)
	}

	return nil
}

type SwitchAction struct {
	SwitchID int64 `json:"id"`
	Status   bool  `json:"status"`
}

func mapSwitchAction(data repository.Switch, status bool) (SwitchAction, error) {
	var switchAction SwitchAction
	switchAction.SwitchID = data.ID
	switchAction.Status = status
	return switchAction, nil
}

func convertDateToUTC(date time.Time) time.Time {
	// Convert date to UTC
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		fmt.Println("Error loading UTC location:", err)
	}
	return date.In(utc)
}
