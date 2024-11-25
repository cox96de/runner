package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/api/httpserverclient"
	"github.com/cox96de/runner/log"
	"github.com/samber/lo"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var greenColor = color.New(color.FgGreen).SprintFunc()

func main() {
	serverAddr := ""
	flagSet := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	flagSet.StringVarP(&serverAddr, "server", "s", "", "server address")
	err := flagSet.Parse(os.Args[1:])
	checkErr(err, "failed to parse flags")
	pipelineFiles := flagSet.Args()
	client := composeClient(serverAddr)
	sig := make(chan os.Signal, 1)
	stop := make(chan struct{}, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		close(stop)
		log.Info("signal received")
		log.Info("try again to force exit")
		<-sig
	}()
	ctx := context.Background()
	for _, file := range pipelineFiles {
		fileContent, err := os.ReadFile(file)
		checkErr(err, "failed to read file", file)
		p := &api.PipelineDSL{}
		err = yaml.Unmarshal(fileContent, p)
		if err != nil {
			checkErr(err)
		}
		pipeline, err := client.CreatePipeline(ctx, &api.CreatePipelineRequest{Pipeline: p})
		checkErr(err, "failed to create pipeline")
		color.Green("########### pipeline '%d' created ###########\n", pipeline.Pipeline.ID)
		for _, job := range pipeline.Pipeline.Jobs {
			if err := watchJob(ctx, client, job, stop); err != nil {
				log.Errorf("%+v", err)
			}
		}
	}
}

func composeClient(serverAddr string) api.ServerClient {
	client, err := httpserverclient.NewClient(&http.Client{}, serverAddr)
	checkErr(err)
	return client
}

func watchJob(ctx context.Context, client api.ServerClient, job *api.Job, stop chan struct{}) error {
	var jobExecution *api.JobExecution
	status := api.StatusCreated
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-stop:
			_, err := client.CancelJobExecution(ctx, &api.CancelJobExecutionRequest{JobExecutionID: jobExecution.ID})
			if err != nil {
				log.Errorf("failed to cancel job execution: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}()
wait:
	for {
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			getJobExecutionResponse, err := client.GetJobExecution(ctx, &api.GetJobExecutionRequest{
				JobExecutionID: job.Execution.ID,
			})
			if err != nil {
				return err // TODO: handle error
			}
			jobExecution = getJobExecutionResponse.JobExecution
			if jobExecution.Status != status {
				color.Green("########### job status transmit to %s ###########\n", jobExecution.Status)
				status = jobExecution.Status
			}
			switch {
			case jobExecution.Status.IsCompleted():
				return nil
			case jobExecution.Status == api.StatusRunning:
				break wait
			}
		}
	}
	// Fetch logs.
	for _, step := range job.Steps {
		err := watchStep(ctx, client, jobExecution, step, stop)
		if err != nil {
			return err
		}
	}
	return nil
}

func watchStep(ctx context.Context, client api.ServerClient, jobExecution *api.JobExecution, step *api.Step, stop chan struct{}) error {
	offset := int64(0)
	for {
		select {
		case <-stop:
			log.Infof("cancel job execution")
			_, err := client.CancelJobExecution(context.Background(),
				&api.CancelJobExecutionRequest{JobExecutionID: jobExecution.ID})
			if err != nil {
				log.Errorf("failed to cancel job execution: %v", err)
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			getLogLinesResponse, err := client.GetLogLines(ctx, &api.GetLogLinesRequest{
				JobExecutionID: jobExecution.ID,
				Name:           step.Name,
				Offset:         offset,
				Limit:          lo.ToPtr(int64(100)),
			})
			if err != nil {
				return err
			}
			for _, logLine := range getLogLinesResponse.Lines {
				fmt.Printf("%s %s\n", greenColor(logLine.Number), logLine.Output)
			}
			if len(getLogLinesResponse.Lines) == 0 {
				getStepExecutionResponse, err := client.GetStepExecution(context.Background(), &api.GetStepExecutionRequest{
					StepExecutionID: step.Execution.ID,
				})
				if err != nil {
					return err
				}
				stepExecution := getStepExecutionResponse.StepExecution
				if stepExecution.Status.IsCompleted() {
					color.Green("########### step '%s' execution '%d' exit with status: %s, exit code: %d ###########\n", step.Name, stepExecution.ID,
						stepExecution.Status, stepExecution.ExitCode)
					return nil
				}
			}
			offset += int64(len(getLogLinesResponse.Lines))
		}
	}
}

func checkErr(err error, msg ...string) {
	if err == nil {
		return
	}
	log.Fatal(msg, err)
}
