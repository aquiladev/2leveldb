package main

var (
	cfg *config
)

func actor(actorChan chan<- *converter) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}
	actorLog.Infof("Config %+v", config)

	cfg = config
	defer func() {
		if logRotator != nil {
			logRotator.Close()
		}
	}()

	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interrupt := interruptListener()
	defer actorLog.Info("Shutdown complete")

	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return nil
	}

	// Create worker and start it.
	converter, err := newConverter(cfg)
	if err != nil {
		actorLog.Errorf("Unable to start converter: %v", err)
		return err
	}
	defer func() {
		actorLog.Infof("Gracefully shutting down the converter...")
		converter.Stop()
		converter.WaitForShutdown()
		actorLog.Infof("Worker shutdown complete")
	}()
	converter.Start()

	if actorChan != nil {
		actorChan <- converter
	}

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil
}