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
		"N1_HIGHCPU_8":  0.016,
		"N1_HIGHCPU_32": 0.064,
		"E2_HIGHCPU_8":  0.016,
		"E2_HIGHCPU_32": 0.064,
	}

	// Set up a channel to receive results from each build
	resultChan := make(chan Result)
	done := make(chan bool)

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

	// Print table of cost
	for _, result := range results {
		fmt.Printf("Build took %v on %v and cost $%.3f\n", result.Duration, result.MachineType, result.Duration.Minutes()*machineTypes[result.MachineType])
	}
}

func runBuild(machineType string, result chan Result, done chan bool) {
	command := "gcloud"
	args := strings.Split(fmt.Sprintf("builds submit . --machine-type=%s", machineType), " ")
	cmd := exec.Command(command, args...)
	// stderr, _ := cmd.StderrPipe()
	// stdout, _ := cmd.StdoutPipe()

	// Start build timer
	startTime := time.Now()

	// Run command
	fmt.Printf("Starting build on %s...\n", machineType)
	cmd.Env = []string{"PYTHONUNBUFFERED=TRUE"}
	cmd.Start()

	// // Get STDOUT and STDERR buffers
	// outscanner := bufio.NewScanner(stdout)
	// errscanner := bufio.NewScanner(stderr)
	// for errscanner.Scan() || outscanner.Scan() {
	// 	o := outscanner.Text()
	// 	fmt.Print(o)
	// 	e := errscanner.Text()
	// 	fmt.Print(e)
	// 	fmt.Println()
	// }
	err := cmd.Wait()
	endTime := time.Now()
	diff := endTime.Sub(startTime)
	fmt.Printf("Build completed on %s in %.2f minutes.\n", machineType, diff.Minutes())

	// Send result of build
	result <- Result{Err: err, Duration: diff, MachineType: machineType}
}
