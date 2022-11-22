package main

import (
	"context"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/aws-asg-cloudfront/cmd/asg-lifecycle-lambda/internal/asglifecycle"
	"github.com/wolfeidau/aws-asg-cloudfront/cmd/asg-lifecycle-lambda/internal/flags"
	"github.com/wolfeidau/lambda-go-extras/lambdaextras"
	"github.com/wolfeidau/lambda-go-extras/middleware"
	"github.com/wolfeidau/lambda-go-extras/standard"
)

var (
	commit = "unknown"

	cli flags.ASGLifecycleLambda
)

func main() {
	kong.Parse(&cli)

	flds := middleware.FieldMap{"commit": commit}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load SDK config")
	}

	asgHandler := asglifecycle.NewAutoscalingHandler(cfg, cli)

	standard.Default(lambdaextras.GenericHandler(asgHandler.Handler), standard.Fields(flds))
}
