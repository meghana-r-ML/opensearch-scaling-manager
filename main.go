package main

import (
	"scaling_manager/cluster"
	"scaling_manager/config"
	"scaling_manager/provision"
	"scaling_manager/task"
	log "scaling_manager/logger"
	"strings"
	"time"
)

var state = new(provision.State)

func main() {
	// A periodic check if there is a change in master node to pick up incomplete provisioning
	go periodicProvisionCheck()
	// The polling interval is set to 5 minutes and can be configured.
	ticker := time.Tick(5 * time.Second)
	for range ticker {
		// This function is responsible for fetching the metrics and pushing it to the index.
		// In starting we will call simulator to provide this details with current timestamp.
		// fetch.FetchMetrics()
		// This function will be responsible for parsing the config file and fill in task_details struct.
		var task = new(task.TaskDetails)
		configStruct := config.GetConfig("config.yaml")
		task.Tasks = configStruct.TaskDetails
		// This function is responsible for evaluating the task and recommend.
		recommendation_list := task.EvaluateTask()
		// This function is responsible for getting the recommendation and provision.
		provision.GetRecommendation(state, recommendation_list)
	}
}

// Input:
// Description: It periodically checks if the master node is changed and picks up if there was any ongoing provision operation
// Output:

func periodicProvisionCheck() {
	tick := time.Tick(5 * time.Second)
	previousMaster := cluster.CheckIfMaster()
	for range tick {
		state.GetCurrentState()
		// Call a function which returns the current master node
		currentMaster := cluster.CheckIfMaster()
		if state.CurrentState != "normal" {
			if !(previousMaster) && currentMaster {
				configStruct, err := config.GetConfig("config.yaml")
				if err != nil {
					log.Warn(log.ProvisionerWarn, "Unable to get Config from GetConfig()")
					return
				}
				cfg := configStruct.ClusterDetails
				if strings.Contains(state.CurrentState, "scaleup") {
					log.Info("Calling scaleOut")
					isScaledUp := provision.ScaleOut(cfg, state)
					if isScaledUp {
						log.Info("Scaleup completed successfully")
					} else {
						// Add a retry mechanism
						log.Warn("Scaleup failed")
					}
				} else if strings.Contains(state.CurrentState, "scaledown") {
					log.Info("Calling scaleIn")
					isScaledDown := provision.ScaleIn(cfg, state)
					if isScaledDown {
						log.Info("Scaledown completed successfully")
					} else {
						// Add a retry mechanism
						log.Warn("Scaledown failed")
					}
				}
			}
		}
		// Update the previousMaster for next loop
		previousMaster = currentMaster
	}
}

