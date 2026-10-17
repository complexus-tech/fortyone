import argparse
import copy
import json
import tempfile
import unittest
from pathlib import Path

import ecs_release as release


DEFINITION = {
    "family": "fortyone-api", "networkMode": "awsvpc", "requiresCompatibilities": ["FARGATE"],
    "taskRoleArn": "task-role", "executionRoleArn": "execution-role", "cpu": "256", "memory": "512",
    "containerDefinitions": [{"name": "api", "image": "server:commit", "essential": True,
                              "healthCheck": {"command": ["CMD-SHELL", "curl localhost/health"]},
                              "environment": [{"name": "EXAMPLE", "value": "preserved"}]}],
    "status": "ACTIVE", "revision": 5, "registeredAt": "ignored",
}
SERVICE = {
    "status": "ACTIVE", "desiredCount": 2, "runningCount": 2, "pendingCount": 0,
    "taskDefinition": "candidate-arn", "deploymentController": {"type": "ECS"},
    "networkConfiguration": {"awsvpcConfiguration": {"subnets": ["private-subnet"], "securityGroups": ["api-group"], "assignPublicIp": "DISABLED"}},
    "deployments": [{"status": "PRIMARY", "rolloutState": "COMPLETED"}],
}


class FakeECS:
    def __init__(self, exit_code=0, running=False):
        self.calls = []
        self.service = copy.deepcopy(SERVICE)
        self.exit_code = exit_code
        self.running = running
        self.launch_failure = False
        self.image = "server:commit"

    def __call__(self, operation, request):
        self.calls.append((operation, request))
        if operation == "describe-services":
            return {"services": [self.service]}
        if operation == "register-task-definition":
            return {"taskDefinition": {"taskDefinitionArn": "schema-arn"}}
        if operation == "run-task":
            if self.launch_failure:
                return {"failures": [{"reason": "capacity"}], "tasks": []}
            return {"tasks": [{"taskArn": "schema-task"}]}
        if operation == "describe-tasks":
            return {"tasks": [{"lastStatus": "RUNNING" if self.running else "STOPPED", "containers": [{"name": "api", "exitCode": self.exit_code}]}]}
        if operation == "describe-task-definition":
            return {"taskDefinition": {"containerDefinitions": [{"name": "api", "image": self.image}]}}
        return {}


class ReleaseTests(unittest.TestCase):
    def setUp(self):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        filename = Path(directory.name) / "task.json"
        filename.write_text(json.dumps(DEFINITION))
        self.args = argparse.Namespace(
            cluster="cluster", service="api-service", container="api", task_definition=str(filename),
            operation="check-schema", execute_reviewed_migrations=False,
            run_id="run-123", timeout=60, expected_image="server:commit",
        )

    def test_readonly_preflight_preserves_role_network_secrets_and_uses_candidate(self):
        ecs = FakeECS()
        release.run_schema(self.args, ecs)
        request = next(body for operation, body in ecs.calls if operation == "register-task-definition")
        self.assertEqual(request["taskRoleArn"], "task-role")
        container = request["containerDefinitions"][0]
        self.assertEqual(container["image"], "server:commit")
        self.assertEqual(container["environment"], DEFINITION["containerDefinitions"][0]["environment"])
        self.assertEqual(container["entryPoint"], ["/app/api"])
        self.assertEqual(container["command"], ["-check-schema"])
        self.assertNotIn("healthCheck", container)
        self.assertNotIn("revision", request)
        launch = next(body for operation, body in ecs.calls if operation == "run-task")
        self.assertEqual(launch["networkConfiguration"], SERVICE["networkConfiguration"])
        self.assertEqual(launch["launchType"], "FARGATE")
        self.assertEqual(ecs.calls[-1][0], "deregister-task-definition")
        self.assertNotIn("update-service", [operation for operation, _ in ecs.calls])

    def test_nonzero_or_absent_exit_status_blocks_release(self):
        for code in (1, 137, None):
            with self.subTest(code=code), self.assertRaises(release.ReleaseError):
                release.run_schema(self.args, FakeECS(exit_code=code))

    def test_launch_failure_blocks_release(self):
        ecs = FakeECS()
        ecs.launch_failure = True
        with self.assertRaises(release.ReleaseError):
            release.run_schema(self.args, ecs)
        self.assertEqual(ecs.calls[-1][0], "deregister-task-definition")

    def test_timeout_stops_only_its_schema_task(self):
        ecs = FakeECS(running=True)
        ticks = iter([0, 0, 61])
        with self.assertRaisesRegex(release.ReleaseError, "timed out"):
            release.run_schema(self.args, ecs, clock=lambda: next(ticks), sleep=lambda _: None)
        self.assertIn(("stop-task", {"cluster": "cluster", "task": "schema-task", "reason": "Release schema command failed or was cancelled"}), ecs.calls)

    def test_cancellation_stops_schema_task(self):
        ecs = FakeECS(running=True)
        def cancel(_):
            raise KeyboardInterrupt()
        with self.assertRaises(KeyboardInterrupt):
            release.run_schema(self.args, ecs, sleep=cancel)
        self.assertEqual([name for name, _ in ecs.calls][-2:], ["stop-task", "deregister-task-definition"])

    def test_stop_failure_still_deregisters_revision_and_blocks_release(self):
        ecs = FakeECS(running=True)
        def fail_stop(operation, request):
            result = ecs(operation, request)
            if operation == "stop-task":
                raise release.ReleaseError("Stop request failed")
            return result
        ticks = iter([0, 0, 61])
        with self.assertRaisesRegex(release.ReleaseError, "Stop request failed"):
            release.run_schema(self.args, fail_stop, clock=lambda: next(ticks), sleep=lambda _: None)
        self.assertEqual([name for name, _ in ecs.calls][-2:], ["stop-task", "deregister-task-definition"])

    def test_manual_migration_requires_explicit_review_flag_before_aws(self):
        self.args.operation = "migrate"
        ecs = FakeECS()
        with self.assertRaises(release.ReleaseError):
            release.run_schema(self.args, ecs)
        self.assertEqual(ecs.calls, [])
        self.args.execute_reviewed_migrations = True
        release.run_schema(self.args, ecs)
        request = next(body for operation, body in ecs.calls if operation == "register-task-definition")
        self.assertEqual(request["containerDefinitions"][0]["command"], ["-migrate"])

    def test_zero_desired_tasks_cannot_pass_preflight_or_verification(self):
        ecs = FakeECS()
        ecs.service["desiredCount"] = 0
        with self.assertRaises(release.ReleaseError):
            release.run_schema(self.args, ecs)
        with self.assertRaises(release.ReleaseError):
            release.verify_release(self.args, ecs)

    def test_healthy_candidate_passes_verification(self):
        release.verify_release(self.args, FakeECS())

    def test_completed_rollback_is_not_a_successful_release(self):
        ecs = FakeECS()
        ecs.image = "server:old"
        with self.assertRaisesRegex(release.ReleaseError, "rolled back"):
            release.verify_release(self.args, ecs)

    def test_incomplete_and_unhealthy_deployments_are_rejected(self):
        for field, value in (("runningCount", 1), ("pendingCount", 1), ("deployments", [{"status": "PRIMARY", "rolloutState": "IN_PROGRESS"}])):
            ecs = FakeECS()
            ecs.service[field] = value
            with self.subTest(field=field), self.assertRaises(release.ReleaseError):
                release.verify_release(self.args, ecs)

    def test_workflow_checks_schema_before_api_and_waits_for_both_services(self):
        workflow = (Path(__file__).parents[1] / "workflows/ecs-fargate-release.yml").read_text()
        deploy = workflow.split("\n  deploy:\n", 1)[1]
        positions = [deploy.index(name) for name in (
            "name: Verify candidate database schema", "name: Submit server rollout",
            "name: Verify server candidate is healthy", "name: Submit worker rollout",
            "name: Verify worker candidate is healthy",
        )]
        self.assertEqual(positions, sorted(positions))
        self.assertEqual(deploy.count("wait-for-service-stability: true"), 2)
        self.assertIn("ecs_release.py check-schema", deploy)
        self.assertNotIn("ecs_release.py migrate", deploy)
        self.assertNotIn("execute-reviewed-migrations", deploy)


if __name__ == "__main__":
    unittest.main()
