#!/usr/bin/env python3
"""Run an isolated schema command, or verify a completed ECS release.

The normal release uses check-schema only. Applying migrations requires the
separate explicit operator flag and never updates either ECS service.
"""

import argparse
import copy
import hashlib
import json
import signal
import subprocess
import tempfile
import time
from pathlib import Path


class ReleaseError(RuntimeError):
    pass


def aws(operation, request):
    # Task definitions may contain sensitive environment settings. Never log the
    # request or place its JSON in process arguments or workflow outputs.
    with tempfile.TemporaryDirectory(prefix="ecs-release-") as directory:
        filename = Path(directory) / "request.json"
        filename.write_text(json.dumps(request))
        filename.chmod(0o600)
        result = subprocess.run(
            ["aws", "ecs", operation, "--cli-input-json", f"file://{filename}",
             "--output", "json", "--no-cli-pager"],
            capture_output=True, text=True, timeout=70, check=False,
        )
    if result.returncode:
        raise ReleaseError(f"ECS {operation} failed (exit {result.returncode}); inspect AWS access/configuration")
    return json.loads(result.stdout or "{}")


def service_state(cluster, service, call=aws):
    response = call("describe-services", {"cluster": cluster, "services": [service]})
    if response.get("failures") or len(response.get("services", [])) != 1:
        raise ReleaseError("ECS service could not be resolved uniquely")
    current = response["services"][0]
    if current.get("status") != "ACTIVE" or current.get("desiredCount", 0) < 1:
        raise ReleaseError("ECS service must be active with at least one desired task")
    if current.get("deploymentController", {}).get("type", "ECS") != "ECS":
        raise ReleaseError("Only rolling ECS services are supported by this release helper")
    return current


def schema_definition(source, container_name, operation):
    allowed = (
        "family", "taskRoleArn", "executionRoleArn", "networkMode", "containerDefinitions",
        "volumes", "placementConstraints", "requiresCompatibilities", "cpu", "memory",
        "pidMode", "ipcMode", "proxyConfiguration", "inferenceAccelerators",
        "ephemeralStorage", "runtimePlatform", "enableFaultInjection",
    )
    definition = {key: copy.deepcopy(source[key]) for key in allowed if key in source}
    if definition.get("networkMode") != "awsvpc" or "FARGATE" not in definition.get("requiresCompatibilities", []):
        raise ReleaseError("Schema commands require the reviewed Fargate/awsvpc task definition")
    containers = [item for item in definition["containerDefinitions"] if item["name"] == container_name]
    if len(containers) != 1:
        raise ReleaseError("API container is missing or ambiguous")
    container = containers[0]
    container.pop("healthCheck", None)  # No HTTP server runs in a schema task.
    container["entryPoint"] = ["/app/api"]
    container["command"] = [f"-{operation}"]
    container["essential"] = True
    definition["family"] = definition["family"][:240] + "-schema"
    return definition


def run_schema(args, call=aws, clock=time.monotonic, sleep=time.sleep):
    if args.operation == "migrate" and not args.execute_reviewed_migrations:
        raise ReleaseError("Migration requires --execute-reviewed-migrations after backup and rollout review")
    current = service_state(args.cluster, args.service, call)
    network = current.get("networkConfiguration", {}).get("awsvpcConfiguration")
    if not network or not network.get("subnets"):
        raise ReleaseError("API service has no usable awsvpc network configuration")
    source = json.loads(Path(args.task_definition).read_text())
    definition = schema_definition(source, args.container, args.operation)
    registered = call("register-task-definition", definition)
    definition_arn = registered["taskDefinition"]["taskDefinitionArn"]
    task_arn = None
    stopped = False
    try:
        request = {
            "cluster": args.cluster, "taskDefinition": definition_arn, "count": 1,
            "launchType": "FARGATE", "networkConfiguration": {"awsvpcConfiguration": network},
            "clientToken": hashlib.sha256(f"{args.run_id}:{definition_arn}:{args.operation}".encode()).hexdigest(),
            "startedBy": f"release-{args.operation}",
        }
        if current.get("platformVersion"):
            request["platformVersion"] = current["platformVersion"]
        response = call("run-task", request)
        launched = response.get("tasks", [])
        if launched:
            task_arn = launched[0]["taskArn"]
        if response.get("failures") or len(launched) != 1:
            raise ReleaseError("ECS could not start exactly one schema task")
        print(f"Schema task: {task_arn}", flush=True)
        deadline = clock() + args.timeout
        while clock() < deadline:
            response = call("describe-tasks", {"cluster": args.cluster, "tasks": [task_arn]})
            if response.get("failures") or len(response.get("tasks", [])) != 1:
                raise ReleaseError("Schema task state is unavailable; refusing release")
            task = response["tasks"][0]
            if task.get("lastStatus") == "STOPPED":
                stopped = True
                containers = [item for item in task.get("containers", []) if item["name"] == args.container]
                if len(containers) != 1 or containers[0].get("exitCode") != 0:
                    raise ReleaseError(f"Schema {args.operation} failed; inspect the task's CloudWatch logs")
                print(f"Schema {args.operation} passed", flush=True)
                return
            sleep(5)
        raise ReleaseError("Schema task timed out; release is blocked")
    finally:
        try:
            if task_arn and not stopped:
                call("stop-task", {"cluster": args.cluster, "task": task_arn, "reason": "Release schema command failed or was cancelled"})
        finally:
            # Even a failed stop request must not skip revision cleanup. A stop
            # failure still propagates, so an uncertain task never permits release.
            call("deregister-task-definition", {"taskDefinition": definition_arn})


def verify_release(args, call=aws):
    current = service_state(args.cluster, args.service, call)
    deployments = current.get("deployments", [])
    if (len(deployments) != 1 or deployments[0].get("status") != "PRIMARY"
            or deployments[0].get("rolloutState") != "COMPLETED"
            or current.get("pendingCount") != 0
            or current.get("runningCount", 0) != current["desiredCount"]):
        raise ReleaseError("Service has not completed a healthy rollout")
    definition = call("describe-task-definition", {"taskDefinition": current["taskDefinition"]})["taskDefinition"]
    containers = [item for item in definition["containerDefinitions"] if item["name"] == args.container]
    if len(containers) != 1 or containers[0].get("image") != args.expected_image:
        raise ReleaseError("Service is not running the candidate image; it may have rolled back")
    print(f"Verified {args.service} candidate rollout", flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("operation", choices=("check-schema", "migrate", "verify"))
    parser.add_argument("--cluster", required=True)
    parser.add_argument("--service", required=True)
    parser.add_argument("--container", required=True)
    parser.add_argument("--task-definition")
    parser.add_argument("--expected-image")
    parser.add_argument("--run-id")
    parser.add_argument("--timeout", type=int, default=300)
    parser.add_argument("--execute-reviewed-migrations", action="store_true")
    args = parser.parse_args()
    if args.operation == "verify":
        if not args.expected_image:
            parser.error("verify requires --expected-image")
        verify_release(args)
        return
    if not args.task_definition or not args.run_id or not 30 <= args.timeout <= 3600:
        parser.error("schema commands require --task-definition, --run-id, and a timeout of 30–3600 seconds")
    run_schema(args)


def interrupt_release(_signum, _frame):
    raise KeyboardInterrupt()


if __name__ == "__main__":
    signal.signal(signal.SIGTERM, interrupt_release)
    try:
        main()
    except (ReleaseError, subprocess.TimeoutExpired, KeyboardInterrupt) as error:
        raise SystemExit(f"Release stopped: {error}") from None
