#!/usr/bin/env node

import { execFile as execFileCallback } from "node:child_process";
import { promises as fs } from "node:fs";
import path from "node:path";
import { promisify } from "node:util";
import { fileURLToPath } from "node:url";
import {
  compareBaselineTransitions,
  findApprovedPolicyTransition,
} from "./architecture.mjs";

const execFile = promisify(execFileCallback);
const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const appRoot = path.resolve(scriptDirectory, "../..");
const baselinePath = path.join(appRoot, "architecture-debt-baseline.json");
const baselineRepositoryPath = "apps/projects/architecture-debt-baseline.json";

const usage = `Usage:
  node scripts/architecture/verify-baseline.mjs --base-ref <merge-base>`;

const parseArguments = (arguments_) => {
  let baseRef = null;

  for (let index = 0; index < arguments_.length; index += 1) {
    const argument = arguments_[index];
    if (argument === "--base-ref") {
      baseRef = arguments_[index + 1]?.trim() ?? "";
      index += 1;
      if (!baseRef || baseRef.startsWith("-")) {
        throw new Error("--base-ref requires a commit or ref");
      }
      continue;
    }
    if (argument === "--help" || argument === "-h") return { help: true };
    throw new Error(`Unknown argument: ${argument}`);
  }

  if (!baseRef) throw new Error("--base-ref is required");
  return { baseRef, help: false };
};

const readJson = (source, description) => {
  try {
    return JSON.parse(source);
  } catch (error) {
    throw new Error(
      `${description} is not valid JSON: ${error instanceof Error ? error.message : String(error)}`,
    );
  }
};

const formatIssue = (issue) => {
  const change =
    issue.type === "new"
      ? `new count ${issue.currentCount}`
      : `${issue.baselineCount} -> ${issue.currentCount}`;
  return `  [${issue.category}] ${change}: ${issue.key}`;
};

const main = async () => {
  let options;
  try {
    options = parseArguments(process.argv.slice(2));
  } catch (error) {
    process.stderr.write(`${error.message}\n\n${usage}\n`);
    process.exitCode = 2;
    return;
  }

  if (options.help) {
    process.stdout.write(`${usage}\n`);
    return;
  }

  const { stdout: repositoryRootOutput } = await execFile(
    "git",
    ["rev-parse", "--show-toplevel"],
    { cwd: appRoot },
  );
  const repositoryRoot = repositoryRootOutput.trim();
  const { stdout: baseCommitOutput } = await execFile(
    "git",
    ["rev-parse", "--verify", `${options.baseRef}^{commit}`],
    { cwd: repositoryRoot },
  );
  const baseCommit = baseCommitOutput.trim();

  let baseSource;
  try {
    ({ stdout: baseSource } = await execFile(
      "git",
      ["show", `${baseCommit}:${baselineRepositoryPath}`],
      { cwd: repositoryRoot },
    ));
  } catch {
    throw new Error(
      `Architecture baseline is missing at ${baselineRepositoryPath} in merge base ${baseCommit}`,
    );
  }

  const base = readJson(baseSource, "Merge-base architecture baseline");
  const candidate = readJson(
    await fs.readFile(baselinePath, "utf8"),
    "Current architecture baseline",
  );
  const issues = compareBaselineTransitions(base, candidate);

  if (issues.length > 0) {
    process.stderr.write(
      `Architecture baseline weakens compared with merge base ${baseCommit}:\n${issues
        .map(formatIssue)
        .join("\n")}\n`,
    );
    process.exitCode = 1;
    return;
  }

  const policyTransition = findApprovedPolicyTransition(base, candidate);
  if (policyTransition) {
    try {
      const { stdout: objectType } = await execFile(
        "git",
        ["cat-file", "-t", `${baseCommit}:${policyTransition.reference}`],
        { cwd: repositoryRoot },
      );
      if (objectType.trim() !== "blob") {
        throw new Error("reference is not a file");
      }
    } catch {
      throw new Error(
        `Approved policy transition reference must exist in the merge base: ${policyTransition.reference}`,
      );
    }
  }

  process.stdout.write(
    `Architecture baseline integrity passed against merge base ${baseCommit}.\n`,
  );
};

await main().catch((error) => {
  process.stderr.write(`${error.stack ?? error.message}\n`);
  process.exitCode = 1;
});
