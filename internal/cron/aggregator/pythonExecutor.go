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

func (a *aggregator) finalize(ctx context.Context) ([]byte, error) {
	globalWeights, err := runPythonFedAvg(ctx, fedAvgPayload{Entries: a.entries, TotalExamples: a.totalExamples})
	if err != nil {
		return nil, err
	}

	return globalWeights, nil
}

func runPythonFedAvg(
	ctx context.Context,
	payload fedAvgPayload,
) ([]byte, error) {

	cmd := exec.CommandContext(
		ctx,
		"python3",
		"/app/fedavg.py",
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

	// --- writer goroutine ---
	go func() {
		defer stdin.Close()

		// простой бинарный протокол:
		// [total_examples][num_entries]
		// повтор:
		//   [num_examples][weights_len][weights_bytes]

		bw := bufio.NewWriter(stdin)
		defer bw.Flush()

		_ = binary.Write(bw, binary.LittleEndian, payload.TotalExamples)
		_ = binary.Write(bw, binary.LittleEndian, uint64(len(payload.Entries)))

		for _, e := range payload.Entries {
			_ = binary.Write(bw, binary.LittleEndian, e.numExamples)
			_ = binary.Write(bw, binary.LittleEndian, uint64(len(e.weights)))
			_, _ = bw.Write(e.weights)
		}
	}()

	outBytes, err := io.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	errBytes, _ := io.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf(
			"python fedavg failed: %w, stderr=%s",
			err,
			string(errBytes),
		)
	}

	return outBytes, nil
}
