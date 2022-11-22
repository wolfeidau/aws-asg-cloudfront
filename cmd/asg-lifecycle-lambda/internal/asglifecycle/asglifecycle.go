package asglifecycle

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/aws-asg-cloudfront/cmd/asg-lifecycle-lambda/internal/events/ec2instance"
	"github.com/wolfeidau/aws-asg-cloudfront/cmd/asg-lifecycle-lambda/internal/flags"
)

type AutoscalingHandler struct {
	autoscaling *autoscaling.Client
	route53     *route53.Client
	ec2         *ec2.Client
	cli         flags.ASGLifecycleLambda
}

func NewAutoscalingHandler(cfg aws.Config, cli flags.ASGLifecycleLambda) *AutoscalingHandler {
	return &AutoscalingHandler{
		autoscaling: autoscaling.NewFromConfig(cfg),
		route53:     route53.NewFromConfig(cfg),
		ec2:         ec2.NewFromConfig(cfg),
		cli:         cli,
	}
}

func (asgh *AutoscalingHandler) updateRoute53(ctx context.Context, action types.ChangeAction, ec2InstanceId, ip, hostname string) error {
	recordSet := &types.ResourceRecordSet{
		Name:             aws.String(hostname),
		Type:             types.RRTypeA,
		MultiValueAnswer: aws.Bool(true),
		SetIdentifier:    aws.String(ec2InstanceId),
		TTL:              aws.Int64(300),
		ResourceRecords: []types.ResourceRecord{
			{
				Value: aws.String(ip),
			},
		},
	}

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action:            action,
					ResourceRecordSet: recordSet,
				},
			},
		},
		HostedZoneId: &asgh.cli.HostedZoneID,
	}

	changeResult, err := asgh.route53.ChangeResourceRecordSets(ctx, params)
	if err != nil {
		return err
	}

	log.Ctx(ctx).Info().Fields(map[string]any{
		"request": params,
		"result":  changeResult,
	}).Msg("complete route53 changeset")

	return nil
}

func (asgh *AutoscalingHandler) getASGTags(ctx context.Context, autoScalingGroupName string) (map[string]string, error) {

	tags := make(map[string]string)

	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{autoScalingGroupName},
	}

	asgList, err := asgh.autoscaling.DescribeAutoScalingGroups(ctx, params)
	if err != nil {
		return nil, err
	}

	log.Ctx(ctx).Info().Fields(map[string]any{
		"request": params,
		"result":  asgList,
	}).Msg("complete describe autoscaling groups")

	for _, t := range asgList.AutoScalingGroups[0].Tags {
		tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}

	return tags, nil
}

func (asgh *AutoscalingHandler) getEc2IP(ctx context.Context, ec2InstanceId string) (string, error) {
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []string{ec2InstanceId},
	}

	ec2Result, err := asgh.ec2.DescribeInstances(ctx, params)
	if err != nil {
		return "", err
	}

	log.Ctx(ctx).Info().Fields(map[string]any{
		"request": params,
		"result":  ec2Result,
	}).Msg("complete ec2 describe instance")

	return aws.ToString(ec2Result.Reservations[0].Instances[0].PublicIpAddress), nil
}

func (asgh *AutoscalingHandler) getRoute53IP(ctx context.Context, ec2InstanceId, hostname string) (string, error) {

	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId:          aws.String(asgh.cli.HostedZoneID),
		StartRecordName:       aws.String(hostname),
		StartRecordIdentifier: aws.String(ec2InstanceId),
		StartRecordType:       types.RRTypeA,
		MaxItems:              aws.Int32(1),
	}

	listResult, err := asgh.route53.ListResourceRecordSets(ctx, params)
	if err != nil {
		return "", err
	}

	log.Ctx(ctx).Info().Fields(map[string]any{
		"request": params,
		"result":  listResult,
	}).Msg("complete route53 list resource record sets")

	return aws.ToString(listResult.ResourceRecordSets[0].ResourceRecords[0].Value), nil
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

	asgTags, err := asgh.getASGTags(ctx, evt.Detail.AutoScalingGroupName)
	if err != nil {
		return nil, err
	}

	log.Ctx(ctx).Info().Str("LifecycleTransition", evt.Detail.LifecycleTransition).Str("InternalServiceHostname", asgTags["InternalServiceHostname"]).Msg("process lifecycle transition")

	switch evt.Detail.LifecycleTransition {
	case "autoscaling:EC2_INSTANCE_LAUNCHING":
		instanceIP, err := asgh.getEc2IP(ctx, evt.Detail.EC2InstanceId)
		if err != nil {
			return nil, err
		}

		err = asgh.updateRoute53(ctx, types.ChangeActionUpsert, evt.Detail.EC2InstanceId, instanceIP, asgTags["InternalServiceHostname"])
		if err != nil {
			return nil, err
		}

	case "autoscaling:EC2_INSTANCE_TERMINATING":
		instanceIP, err := asgh.getRoute53IP(ctx, evt.Detail.EC2InstanceId, asgTags["InternalServiceHostname"])
		if err != nil {
			return nil, err
		}

		err = asgh.updateRoute53(ctx, types.ChangeActionDelete, evt.Detail.EC2InstanceId, instanceIP, asgTags["InternalServiceHostname"])
		if err != nil {
			return nil, err
		}
	}

	err = asgh.sendLifecycleEvent(ctx, evt, "CONTINUE")
	if err != nil {
		return nil, err
	}

	return evt, nil
}
