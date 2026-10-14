import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { pathToFileURL } from "node:url";

const PHASES = ["idle", "workload", "page-ticket", "idle-after"];

function distribution(values) {
  if (!values.length) return null;
  const sorted = [...values].sort((a, b) => a - b);
  const middle = Math.floor(sorted.length / 2);
  return {
    median:
      sorted.length % 2
        ? sorted[middle]
        : (sorted[middle - 1] + sorted[middle]) / 2,
    min: sorted[0],
    max: sorted.at(-1),
  };
}

export function analyzeRun(run) {
  if (run.schemaVersion !== 1) throw new Error("Unsupported benchmark schema.");
  const timebase = run.metadata.timebase;
  const cpuScale =
    run.cpuEncoding === "milliseconds"
      ? 1
      : run.cpuEncoding === "machTicksPerMillion" && timebase?.denom > 0
        ? timebase.numer / timebase.denom
        : null;
  if (cpuScale === null) throw new Error("Missing CPU units or Mach timebase.");

  const trials = new Map(
    run.plan.map((trial) => [trial.id, { config: trial, phases: {} }]),
  );
  for (const event of run.events) {
    const trial = trials.get(event.id);
    if (!trial) continue;
    if (event.type === "ready") trial.initial = event.initial;
    if (event.type === "phase-end") {
      trial.phases[event.result.name] = {
        ...event.result,
        webContentCpuMs:
          typeof event.result.webContentCpuMs === "number"
            ? event.result.webContentCpuMs * cpuScale
            : null,
      };
    }
    if (event.type === "trial-end") {
      trial.complete = true;
      trial.errors = event.errors;
    }
    if (event.type === "fatal") trial.fatal = event.error;
  }

  for (const trial of trials.values()) {
    const checkpoints = [trial.initial, ...Object.values(trial.phases)];
    trial.exclusionReasons = [];
    if (
      !trial.complete ||
      !trial.initial ||
      PHASES.some((p) => !trial.phases[p])
    ) {
      trial.exclusionReasons.push("incomplete");
    }
    if (trial.fatal || trial.errors?.length)
      trial.exclusionReasons.push("error");
    if (
      checkpoints.some(
        (point) => point?.visibility !== "visible" || point?.hasFocus !== true,
      )
    ) {
      trial.exclusionReasons.push("not visible/focused at every checkpoint");
    }
    if (trial.config.mode === "on" && !trial.initial?.controllerActive) {
      trial.exclusionReasons.push("fallback was not active");
    }
  }

  const groups = [];
  const acceptedPairs = [];
  const excludedPairs = [];
  const pairKeys = new Set(
    run.plan
      .filter((trial) => !trial.warmup)
      .map((t) => `${t.cards}-${t.pair}`),
  );
  for (const key of pairKeys) {
    const off = trials.get(`${key}-off`);
    const on = trials.get(`${key}-on`);
    if (
      !off ||
      !on ||
      off.exclusionReasons.length ||
      on.exclusionReasons.length
    ) {
      excludedPairs.push({
        id: key,
        off: off?.exclusionReasons ?? ["missing"],
        on: on?.exclusionReasons ?? ["missing"],
      });
    } else {
      acceptedPairs.push({ id: key, off, on });
    }
  }

  for (const cards of new Set(run.plan.map((trial) => trial.cards))) {
    const pairs = acceptedPairs.filter(
      (pair) => pair.off.config.cards === cards,
    );
    if (!pairs.length) continue;
    for (const phase of PHASES) {
      const group = {
        cards,
        targets: distribution(pairs.map((p) => p.on.initial.targets)),
        phase,
        pairIds: pairs.map((p) => p.id),
        off: {},
        on: {},
        paired: {},
      };
      const metrics = {
        durationMs: (value) => value.duration,
        cpuMs: (value) => value.webContentCpuMs,
        p50GapMs: (value) => value.frames.p50,
        p95GapMs: (value) => value.frames.p95,
        maxGapMs: (value) => value.frames.max,
        gapsOver50Ms: (value) => value.frames.over50,
        styleMutations: (value) => value.styleMutations,
      };
      for (const mode of ["off", "on"]) {
        for (const [metric, getValue] of Object.entries(metrics)) {
          group[mode][metric] = distribution(
            pairs
              .map((p) => getValue(p[mode].phases[phase]))
              .filter(Number.isFinite),
          );
        }
      }
      for (const [metric, key] of [
        ["durationMs", "duration"],
        ["cpuMs", "webContentCpuMs"],
      ]) {
        const comparable = pairs.filter(
          (p) =>
            Number.isFinite(p.off.phases[phase][key]) &&
            Number.isFinite(p.on.phases[phase][key]),
        );
        group.paired[`${metric}Delta`] = distribution(
          comparable.map(
            (p) => p.on.phases[phase][key] - p.off.phases[phase][key],
          ),
        );
        group.paired[`${metric}Ratio`] = distribution(
          comparable
            .filter((p) => p.off.phases[phase][key] > 0)
            .map((p) => p.on.phases[phase][key] / p.off.phases[phase][key]),
        );
      }
      groups.push(group);
    }
  }

  return {
    acceptedPairs: acceptedPairs.map((p) => p.id),
    excludedPairs,
    discardedWarmups: run.plan.filter((t) => t.warmup).map((t) => t.id),
    groups,
  };
}

if (
  process.argv[1] &&
  import.meta.url === pathToFileURL(resolve(process.argv[1])).href
) {
  if (!process.argv[2]) {
    console.error("Usage: node analyze.mjs path/to/results.json");
    process.exitCode = 1;
  } else {
    const run = JSON.parse(await readFile(process.argv[2], "utf8"));
    console.log(JSON.stringify(analyzeRun(run), null, 2));
  }
}
