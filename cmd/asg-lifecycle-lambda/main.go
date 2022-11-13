package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/aws-asg-cloudfront/cmd/asg-lifecycle-lambda/internal/events/ec2instance"
	"github.com/wolfeidau/lambda-go-extras/lambdaextras"
	"github.com/wolfeidau/lambda-go-extras/standard"
)

func main() {

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load SDK config")
	}

	asgHandler := NewAutoscalingHandler(cfg)

	standard.Default(lambdaextras.GenericHandler(asgHandler.Handler))
}

type AutoscalingHandler struct {
	autoscaling *autoscaling.Client
}

func NewAutoscalingHandler(cfg aws.Config) *AutoscalingHandler {
	return &AutoscalingHandler{
		autoscaling: autoscaling.NewFromConfig(cfg),
	}
}

func (asgh *AutoscalingHandler) sendLifecycleEvent(ctx context.Context, evt *ec2instance.AWSEvent, result string) error {

	params := &autoscaling.CompleteLifecycleActionInput{
		LifecycleHookName:     aws.String(evt.Detail.LifecycleHookName),
		AutoScalingGroupName:  aws.String(evt.Detail.AutoScalingGroupName),
		LifecycleActionToken:  aws.String(evt.Detail.LifecycleActionToken),
		LifecycleActionResult: aws.String(result),
		InstanceId:            aws.String(evt.Detail.EC2InstanceId),
	}

	completeResult, err := asgh.autoscaling.CompleteLifecycleAction(ctx, params)
	if err != nil {
		return err
	}

	log.Ctx(ctx).Info().Fields(map[string]any{
		"request": params,
		"result":  completeResult,
	}).Msg("complete action sent")

	return nil
}

func (asgh *AutoscalingHandler) Handler(ctx context.Context, evt *ec2instance.AWSEvent) (*ec2instance.AWSEvent, error) {

	err := asgh.sendLifecycleEvent(ctx, evt, "CONTINUE")
	if err != nil {
		return nil, err
	}

	return evt, nil
}
