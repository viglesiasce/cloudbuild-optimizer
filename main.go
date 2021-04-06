package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Result struct {
	Err         error
	Duration    time.Duration
	MachineType string
}

func main() {
	// Mapping of machine types to their price per minute
	machineTypes := map[string]float64{
		"DEFAULT":       0.003,
		"E2_HIGHCPU_8":  0.016,
		"E2_HIGHCPU_32": 0.064,
	}

	// Set up a channel to receive results from each build
	resultChan := make(chan Result)
	done := make(chan bool)

	// View builds in console
	qs := ""
	if arg, err := getProject(); err == nil && arg != "" {
		qs += "?project=" + arg
	}
	fmt.Printf("View your builds here: https://console.cloud.google.com/cloud-build/builds%s\n\n", qs)

	// Run cloud build for each machine type
	for mt := range machineTypes {
		go runBuild(mt, resultChan, done)
	}

	results := []Result{}
	for i := 0; i < len(machineTypes); i++ {
		result := <-resultChan
		results = append(results, result)
	}
	close(resultChan)

	// Print table of costs
	fmt.Println()
	for _, result := range results {
		fmt.Printf("Build took %v minutes on %v and cost $%.3f\n", result.Duration, result.MachineType, result.Duration.Minutes()*machineTypes[result.MachineType])
	}
}

func runBuild(machineType string, result chan Result, done chan bool) {
	command := "gcloud"

	args := strings.Split("builds submit .", " ")
	if machineType != "DEFAULT" {
		args = append(args, fmt.Sprintf("--machine-type=%s", machineType))
	}
	cmd := exec.Command(command, args...)
	// Start build timer
	startTime := time.Now()

	// Run command
	fmt.Printf("Starting build on %s...\n", machineType)
	cmd.Env = []string{"PYTHONUNBUFFERED=TRUE"}
	cmd.Start()

	err := cmd.Wait()
	endTime := time.Now()
	diff := endTime.Sub(startTime)
	fmt.Printf("Build completed on %s in %.2f minutes.\n", machineType, diff.Minutes())

	// Send result of build
	result <- Result{Err: err, Duration: diff, MachineType: machineType}
}

func getProject() (string, error) {
	args := strings.Split("config get-value project", " ")
	cmd := exec.Command("gcloud", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
		return "", err
	}
	return string(output), err
}
