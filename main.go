package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/naughtygopher/errors"
	"github.com/naughtygopher/proberesponder"

	"github.com/naughtygopher/goapp/internal/configs"
	"github.com/naughtygopher/goapp/internal/pkg/logger"
	"github.com/naughtygopher/goapp/internal/pkg/sysignals"
)

// recoverer is used for panic recovery of the application (note: this is not for the HTTP/gRPC servers).
// So that even if the main function panics we can produce required logs for troubleshooting
var exitErr error

func recoverer() {
	exitCode := 0
	var exitInfo any
	rec := recover()
	err, _ := rec.(error)
	if err != nil {
		exitCode = 1
		exitInfo = err
	} else if rec != nil {
		exitCode = 2
		exitInfo = rec
	} else if exitErr != nil {
		exitCode = 3
		exitInfo = exitErr
	}

	// exiting after receiving a quit signal can be considered a *clean/successful* exit
	if errors.Is(exitErr, sysignals.ErrSigQuit) {
		exitCode = 0
	}

	ctx := context.Background()
	// logging this because we have info logs saying "listening to" various port numbers
	// based on the server type (gRPC, HTTP etc.). But it's unclear *from the logs*
	// if the server is up and running, if it exits for any reason
	if exitCode == 0 {
		logger.Info(ctx, fmt.Sprintf("shutdown complete: %+v", exitInfo))
	} else {
		logger.Error(ctx, fmt.Sprintf("shutdown complete (exit: %d): %+v", exitCode, exitInfo))
	}

	os.Exit(exitCode)
}

func main() {
	defer recoverer()
	var (
		ctx                 = context.Background()
		fatalErr            = make(chan error, 1)
		shutdownGraceperiod = time.Minute
		probeInterval       = time.Second * 3
		probestatus         = proberesponder.New()
	)

	cfgs, err := configs.New()
	if err != nil {
		panic(errors.Wrap(err))
	}

	logger.UpdateDefaultLogger(logger.New(
		cfgs.AppName, cfgs.AppVersion, 0,
		map[string]string{
			"env": cfgs.Environment.String(),
		}),
	)

	// This needs to remain after log initialisation and before server initialisation.
	ap := startAPM(ctx, cfgs)

	healthResponder, err := startHealthResponder(ctx, probestatus, fatalErr)
	if err != nil {
		panic(err)
	}

	hserver, gserver := start(ctx, probestatus, cfgs, fatalErr)

	// by now all the intended servers, subscribers etc. are up and running.
	probestatus.SetNotStarted(false)
	probestatus.SetNotReady(false)
	probestatus.SetNotLive(false)

	defer shutdown(
		shutdownGraceperiod,
		probeInterval,
		probestatus,
		healthResponder,
		hserver,
		gserver,
		ap,
	)
	exitErr = <-fatalErr
}
