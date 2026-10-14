import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { analyzeRun } from "./analyze.mjs";

const archived = JSON.parse(
  await readFile(new URL("./results/2026-10-14.json", import.meta.url), "utf8"),
);

function singlePair() {
  const run = structuredClone(archived);
  run.plan = run.plan.filter((trial) => trial.id.startsWith("24-0-"));
  run.events = run.events.filter((event) => event.id.startsWith("24-0-"));
  return run;
}

test("retains only matched visible/focused pairs and excludes document warmups", () => {
  const summary = analyzeRun(archived);
  assert.deepEqual(summary.acceptedPairs.sort(), [
    "24-0",
    "24-1",
    "24-3",
    "24-4",
    "240-1",
    "240-3",
    "240-4",
  ]);
  assert.deepEqual(summary.excludedPairs.map((pair) => pair.id).sort(), [
    "24-2",
    "240-0",
    "240-2",
  ]);
  assert.deepEqual(summary.discardedWarmups, ["warmup-off", "warmup-on"]);
  const large = summary.groups.find(
    (group) => group.cards === 240 && group.phase === "page-ticket",
  );
  assert.equal(large.off.p95GapMs.median, 36);
  assert.equal(large.on.p95GapMs.median, 643);
  assert.equal(large.targets.median, 1200);
});

test("normalizes legacy Mach counters without mutating measured inputs", () => {
  const run = singlePair();
  const raw = run.events.find(
    (event) =>
      event.type === "phase-end" &&
      event.id === "24-0-on" &&
      event.result.name === "page-ticket",
  ).result.webContentCpuMs;
  const normalized = analyzeRun(run).groups.find(
    (group) => group.phase === "page-ticket",
  );
  assert.equal(normalized.on.cpuMs.median, (raw * 125) / 3);
  assert.equal(
    run.events.find(
      (event) =>
        event.type === "phase-end" &&
        event.id === "24-0-on" &&
        event.result.name === "page-ticket",
    ).result.webContentCpuMs,
    raw,
  );
});

test("does not apply Mach conversion to future correctly normalized runs", () => {
  const run = singlePair();
  run.cpuEncoding = "milliseconds";
  for (const event of run.events) {
    if (event.type === "phase-end") event.result.webContentCpuMs = 200;
  }
  assert.equal(analyzeRun(run).groups[0].on.cpuMs.median, 200);
});

test("rejects a pair when either arm errors, is incomplete, or never activates", () => {
  for (const reason of ["error", "incomplete", "inactive"]) {
    const run = singlePair();
    if (reason === "error")
      run.events.find(
        (event) => event.type === "trial-end" && event.id === "24-0-on",
      ).errors = ["fixture failed"];
    if (reason === "incomplete")
      run.events = run.events.filter(
        (event) => !(event.type === "trial-end" && event.id === "24-0-on"),
      );
    if (reason === "inactive")
      run.events.find(
        (event) => event.type === "ready" && event.id === "24-0-on",
      ).initial.controllerActive = false;
    assert.deepEqual(analyzeRun(run).acceptedPairs, [], reason);
  }
});

test("reports unavailable CPU measurements as missing rather than zero", () => {
  const run = singlePair();
  for (const event of run.events) {
    if (event.type === "phase-end") delete event.result.webContentCpuMs;
  }
  const summary = analyzeRun(run);
  assert.equal(summary.acceptedPairs.length, 1);
  assert.equal(summary.groups[0].on.cpuMs, null);
  assert.equal(summary.groups[0].paired.cpuMsRatio, null);
});
