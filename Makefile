APPNAME := aws-asg-cloudfront
STAGE ?= dev
BRANCH ?= master

GOLANGCI_VERSION = v1.49.0

GIT_HASH := $(shell git rev-parse --short HEAD)

DEPLOY_CMD = sam deploy

.PHONY: build
build:
	@goreleaser --snapshot --rm-dist

.PHONY: deploy-vpc
deploy-vpc:
	@echo "--- deploy stack $(APPNAME)-vpc-$(STAGE)-$(BRANCH)"
	@$(DEPLOY_CMD) \
		--no-fail-on-empty-changeset \
		--template-file sam/app/vpc-3azs.yaml \
		--capabilities CAPABILITY_IAM \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=deployment" \
		--stack-name $(APPNAME)-vpc-$(STAGE)-$(BRANCH) \
		--parameter-overrides ClassB=123

.PHONY: deploy-nodes
deploy-nodes:
	@echo "--- deploy stack $(APPNAME)-nodes-$(STAGE)-$(BRANCH)"
	$(eval SAM_BUCKET := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/deploy_bucket' --query 'Parameter.Value' --output text))
	$(eval CF_PREFIX_LIST_ID := $(shell aws ec2 describe-managed-prefix-lists --filters Name=owner-id,Values=AWS Name=prefix-list-name,Values=com.amazonaws.global.cloudfront.origin-facing --query 'PrefixLists[0].PrefixListId' --output text))
	$(eval AL2_ARM_AMI_ID := $(shell aws ssm get-parameters --names /aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-arm64-gp2 --query 'Parameters[0].Value' --output text))
	@$(DEPLOY_CMD) \
		--no-fail-on-empty-changeset \
		--template-file sam/app/nodes.yaml \
		--s3-bucket $(SAM_BUCKET) \
		--s3-prefix sam/$(GIT_HASH) \
		--capabilities CAPABILITY_IAM \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=deployment" \
		--stack-name $(APPNAME)-nodes-$(STAGE)-$(BRANCH) \
		--parameter-overrides ParentVPCStack=$(APPNAME)-vpc-$(STAGE)-$(BRANCH) \
			ImageId=$(AL2_ARM_AMI_ID) \
			PrefixListId=$(CF_PREFIX_LIST_ID) \
			ASGDesiredCapacity=$(DESIRED_CAPACITY) \
			HostedZoneName=$(HOSTED_ZONE_NAME) \
			HostedZoneId=$(HOSTED_ZONE_ID) \
			InternalServiceId=$(INTERNAL_SERVICE_ID)
