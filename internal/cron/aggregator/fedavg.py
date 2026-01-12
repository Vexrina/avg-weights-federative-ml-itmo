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
            raise EOFError("unexpected EOF")
        buf += chunk
    return buf


def read_u64() -> int:
    return struct.unpack("<Q", read_exact(8))[0]


def main():
    # ---- read global weights ----
    global_len = read_u64()
    global_bytes = read_exact(global_len)

    global_state: Dict[str, torch.Tensor] = torch.load(
        io.BytesIO(global_bytes),
        map_location="cpu",
    )

    # ---- read header ----
    total_examples = read_u64()
    num_entries = read_u64()

    if total_examples == 0 or num_entries == 0:
        # нечего агрегировать — возвращаем старые веса
        out = io.BytesIO()
        torch.save(global_state, out)
        sys.stdout.buffer.write(out.getvalue())
        return

    delta_global: Dict[str, torch.Tensor] | None = None

    for _ in range(num_entries):
        num_examples = read_u64()
        delta_len = read_u64()
        delta_bytes = read_exact(delta_len)

        if num_examples == 0:
            continue

        delta = torch.load(
            io.BytesIO(delta_bytes),
            map_location="cpu",
        )

        if delta.keys() != global_state.keys():
            raise ValueError("Model keys mismatch")

        weight = num_examples / total_examples

        if delta_global is None:
            delta_global = {
                k: delta[k].mul(weight)
                for k in delta
            }
        else:
            for k in delta_global:
                delta_global[k].add_(delta[k], alpha=weight)

    if delta_global is None:
        out = io.BytesIO()
        torch.save(global_state, out)
        sys.stdout.buffer.write(out.getvalue())
        return

    # ---- apply delta ----
    new_global = {
        k: global_state[k] + delta_global[k]
        for k in global_state
    }

    # ---- serialize result ----
    out = io.BytesIO()
    torch.save(new_global, out)
    sys.stdout.buffer.write(out.getvalue())


if __name__ == "__main__":
    main()
