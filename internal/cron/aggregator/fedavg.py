import sys
import struct
import torch
import io
from typing import Dict


def read_exact(n: int) -> bytes:
    buf = b""
    while len(buf) < n:
        chunk = sys.stdin.buffer.read(n - len(buf))
        if not chunk:
            raise EOFError("unexpected EOF while reading stdin")
        buf += chunk
    return buf


def read_uint64() -> int:
    return struct.unpack("<Q", read_exact(8))[0]


def main():
    # ---- read header ----
    total_examples = read_uint64()
    num_entries = read_uint64()

    # Нечего агрегировать — выходим молча
    if num_entries == 0 or total_examples == 0:
        return

    global_delta: Dict[str, torch.Tensor] | None = None

    for _ in range(num_entries):
        num_examples = read_uint64()
        weights_len = read_uint64()
        weights_bytes = read_exact(weights_len)

        if num_examples == 0:
            continue

        # Δw клиента
        delta = torch.load(
            io.BytesIO(weights_bytes),
            map_location="cpu",
        )

        weight = num_examples / total_examples

        if global_delta is None:
            # первая дельта
            global_delta = {
                k: v.mul(weight)
                for k, v in delta.items()
            }
        else:
            # защита от несовпадающих моделей
            if global_delta.keys() != delta.keys():
                raise ValueError("Model parameter keys mismatch between clients")

            for k in global_delta:
                global_delta[k].add_(delta[k], alpha=weight)

    if global_delta is None:
        return

    # ---- serialize result ----
    out = io.BytesIO()
    torch.save(global_delta, out)
    sys.stdout.buffer.write(out.getvalue())


if __name__ == "__main__":
    main()
