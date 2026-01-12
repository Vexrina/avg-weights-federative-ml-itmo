// pythonExecutor.go
package aggregator

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
)

func (a *aggregator) finalize(ctx context.Context, latestWeight []byte) ([]byte, error) {
	globalWeights, err := runPythonFedAvg(ctx, latestWeight, fedAvgPayload{Entries: a.entries, TotalExamples: a.totalExamples}, a.pythonPath)
	if err != nil {
		return nil, err
	}

	return globalWeights, nil
}

func runPythonFedAvg(
	ctx context.Context,
	latestWeight []byte,
	payload fedAvgPayload,
	pythonPath string,
) ([]byte, error) {

	cmd := exec.CommandContext(
		ctx,
		".venv/bin/python",
		fmt.Sprintf("%sfedavg.py", pythonPath),
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	/*
		OS-pipe имеет ограниченный буфер ~64 KB.
		если не использовать горутину, го будет и писать в stdin, и читать в stdout
		нельзя просто всё записать, потом читать
		потому что мы не можем контролировать, когда питонячий скрипт начнёт писать в stdout.
		он может начать писать:
		- после чтения 10 байт
		- после чтения 1 клиента
		- после агрегации
		- логировать прогресс
	*/
	go func() {
		defer stdin.Close()

		bw := bufio.NewWriter(stdin)
		defer bw.Flush()

		// ---------- 1. global weights ----------
		_ = binary.Write(bw, binary.LittleEndian, uint64(len(latestWeight)))
		if len(latestWeight) > 0 {
			_, _ = bw.Write(latestWeight)
		}

		// ---------- 2. header ----------
		_ = binary.Write(bw, binary.LittleEndian, payload.TotalExamples)
		_ = binary.Write(bw, binary.LittleEndian, uint64(len(payload.Entries)))

		// ---------- 3. entries ----------
		for _, e := range payload.Entries {
			_ = binary.Write(bw, binary.LittleEndian, e.numExamples)
			_ = binary.Write(bw, binary.LittleEndian, uint64(len(e.weights)))
			if len(e.weights) > 0 {
				_, _ = bw.Write(e.weights)
			}
		}
	}()

	outBytes, err := io.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	errBytes, _ := io.ReadAll(stderr)

	if err = cmd.Wait(); err != nil {
		return nil, fmt.Errorf(
			"python fedavg failed: %w, stderr=%s",
			err,
			string(errBytes),
		)
	}

	return outBytes, nil
}
