package main

import (
	"os"
	"runtime/pprof"
	"runtime/trace"
)

func startProfiling() (func(), error) {
	traceFile, err := os.Create("traceRuntime.out")

	if err != nil {
		return nil, err
	}

	if err := trace.Start(traceFile); err != nil {
		return nil, err
	}

	cpuFile, err := os.Create("cpu.pprof")

	if err != nil {
		return nil, err
	}

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		return nil, err
	}

	return func() {
		pprof.StopCPUProfile()
		trace.Stop()
		_ = traceFile.Close()
		_ = cpuFile.Close()
	}, nil
}
