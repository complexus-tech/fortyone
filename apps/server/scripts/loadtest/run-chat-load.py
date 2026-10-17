#!/usr/bin/env python3
"""Run real Messages HTTP/SSE load against disposable loopback-only services."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import socket
import subprocess
import tempfile
import time


def unused_port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def candidate_fingerprint(server):
    """Identify the actual working-tree input, including uncommitted sources."""
    runtime = []
    for directory in ("cmd", "internal", "pkg", "migrations"):
        for path in (server / directory).rglob("*"):
            if path.is_file() and (
                path.suffix == ".sql"
                or (path.suffix == ".go" and not path.name.endswith("_test.go"))
            ):
                runtime.append(path)
    runtime.extend(server / name for name in ("go.mod", "go.sum", "Dockerfile"))
    harness = list((server / "internal/bootstrap/api").glob("chat_load*_test.go"))
    harness.append(Path(__file__).resolve())

    def digest(paths):
        result = hashlib.sha256()
        for path in sorted(paths):
            name, content = str(path.relative_to(server)).encode(), path.read_bytes()
            # Length framing keeps filename/content boundaries unambiguous.
            for value in (name, content):
                result.update(len(value).to_bytes(8, "big"))
                result.update(value)
        return result.hexdigest()

    def git(*args):
        return subprocess.check_output(["git", *args], cwd=server, text=True).strip()

    return {
        "git_head": git("rev-parse", "HEAD"),
        "working_tree_dirty": bool(git("status", "--porcelain", "--untracked-files=normal")),
        "runtime_sha256": digest(runtime),
        "runtime_file_count": len(runtime),
        "harness_sha256": digest(harness),
        "harness_file_count": len(harness),
        "fingerprint_scope": "Non-test Go and SQL under cmd/internal/pkg/migrations; go.mod, go.sum, Dockerfile. Harness: chat_load*_test.go and this runner. SHA256 over sorted relative filenames/content, each length-prefixed with unsigned 8-byte big-endian length.",
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--users", type=int, default=1000)
    parser.add_argument("--hold-seconds", type=int, default=300)
    parser.add_argument("--ramp-seconds", type=int, default=30)
    parser.add_argument("--active-chat-percent", type=int, default=20)
    parser.add_argument("--synchronized-chat-start", action="store_true")
    parser.add_argument("--report", default="/tmp/fortyone-chat-load.json")
    parser.add_argument("--pg-bin", default="/opt/homebrew/opt/postgresql@18/bin")
    args = parser.parse_args()
    if (
        not 10 <= args.users <= 10000
        or args.hold_seconds < 10
        or args.ramp_seconds < 1
    ):
        parser.error("users must be 10..10000, hold >= 10s and ramp >= 1s")
    if not 0 <= args.active_chat_percent <= 100:
        parser.error("active-chat-percent must be 0..100")
    pg = Path(args.pg_bin)
    server = Path(__file__).resolve().parents[2]
    candidate = candidate_fingerprint(server)
    root = Path(tempfile.mkdtemp(prefix="fortyone-chat-load-"))
    postgres_started = False
    redis = None
    try:
        (root / "sockets").mkdir()
        subprocess.run(
            [str(pg / "initdb"), "-D", str(root / "data"), "-U", "chat_load",
             "-A", "trust", "--locale=C", "--encoding=UTF8"],
            check=True,
            stdout=subprocess.DEVNULL,
        )
        pg_port, redis_port = unused_port(), unused_port()
        while redis_port == pg_port:
            redis_port = unused_port()
        subprocess.run(
            [str(pg / "pg_ctl"), "-D", str(root / "data"), "-l",
             str(root / "postgres.log"), "-o",
             f"-h 127.0.0.1 -p {pg_port} -k {root / 'sockets'} -c max_connections=100",
             "-w", "start"],
            check=True,
            stdout=subprocess.DEVNULL,
        )
        postgres_started = True
        redis_log = (root / "redis.log").open("w")
        redis = subprocess.Popen(
            [shutil.which("redis-server") or "redis-server", "--bind", "127.0.0.1",
             "--port", str(redis_port), "--save", "", "--appendonly", "no",
             "--dir", str(root), "--maxclients", "10000"],
            stdout=redis_log,
            stderr=subprocess.STDOUT,
        )
        for _ in range(100):
            try:
                with socket.create_connection(("127.0.0.1", redis_port), timeout=0.1):
                    break
            except OSError:
                time.sleep(0.05)
        else:
            raise RuntimeError("Disposable Redis did not start")
        env = os.environ.copy()
        env.update(
            TEST_DATABASE_URL=f"postgres://chat_load@127.0.0.1:{pg_port}/postgres?sslmode=disable",
            TEST_REDIS_URL=f"redis://127.0.0.1:{redis_port}/0",
            CHAT_LOAD_USERS=str(args.users),
            CHAT_LOAD_HOLD_SECONDS=str(args.hold_seconds),
            CHAT_LOAD_RAMP_SECONDS=str(args.ramp_seconds),
            CHAT_LOAD_ACTIVE_CHAT_PERCENT=str(args.active_chat_percent),
            CHAT_LOAD_SYNCHRONIZED_CHAT_START=str(args.synchronized_chat_start).lower(),
            CHAT_LOAD_REPORT=str(Path(args.report).resolve()),
        )
        result = subprocess.run(
            ["go", "test", "-tags=integration,chatload", "./internal/bootstrap/api",
             "-run", "^TestMessagesConcurrentLoad$", "-count=1", "-timeout=20m", "-v"],
            cwd=server,
            env=env,
        )
        report_path = Path(args.report).resolve()
        if report_path.is_file():
            report = json.loads(report_path.read_text())
            report["candidate"] = candidate
            report_path.write_text(json.dumps(report, indent=2) + "\n")
        return result.returncode
    finally:
        if redis is not None:
            redis.terminate()
            try:
                redis.wait(timeout=15)
            except subprocess.TimeoutExpired:
                redis.kill()
                redis.wait()
        if postgres_started:
            subprocess.run(
                [str(pg / "pg_ctl"), "-D", str(root / "data"), "-m", "fast", "-w", "stop"],
                check=True,
                stdout=subprocess.DEVNULL,
            )
        shutil.rmtree(root)
        print("Disposable PostgreSQL and Redis removed. Existing databases were not used.", flush=True)


if __name__ == "__main__":
    raise SystemExit(main())
