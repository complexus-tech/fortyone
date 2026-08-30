#!/usr/bin/env node

import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  CATEGORY_ORDER,
  compareSnapshots,
  formatSnapshot,
  prepareSnapshotForWrite,
  scanArchitecture,
  validateBaseline,
} from "./architecture.mjs";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const appRoot = path.resolve(scriptDirectory, "../..");
const baselinePath = path.join(appRoot, "architecture-debt-baseline.json");

const usage = `Usage:
  node scripts/architecture/check.mjs
  node scripts/architecture/check.mjs --write-baseline
  node scripts/architecture/check.mjs --write-baseline --force-with-adr <reference>

The ordinary baseline writer accepts debt reductions only. New, grown, or moved
debt requires the deliberately noisy --force-with-adr flag and a reviewable ADR
or architecture-plan reference. A scanner-policy change also requires that
explicit ADR reference and records the policy transition. Oversized-file debt
is a hard ceiling: an ADR cannot approve a new oversized file or growth in an
existing one.`;

const parseArguments = (arguments_) => {
  let writeBaseline = false;
  let forceAdrReference = null;

  for (let index = 0; index < arguments_.length; index += 1) {
    const argument = arguments_[index];
    if (argument === "--write-baseline") {
      writeBaseline = true;
      continue;
    }
    if (argument === "--force-with-adr") {
      forceAdrReference = arguments_[index + 1]?.trim() ?? "";
      index += 1;
      if (!forceAdrReference || forceAdrReference.startsWith("--")) {
        throw new Error("--force-with-adr requires a non-empty reference");
      }
      continue;
    }
    if (argument === "--help" || argument === "-h") {
      return { help: true };
    }
    throw new Error(`Unknown argument: ${argument}`);
  }

  if (forceAdrReference && !writeBaseline) {
    throw new Error("--force-with-adr is valid only with --write-baseline");
  }

  return { forceAdrReference, help: false, writeBaseline };
};

const readBaseline = async ({ allowPolicyMismatch = false, required }) => {
  try {
    const baseline = JSON.parse(await fs.readFile(baselinePath, "utf8"));
    validateBaseline(baseline, { allowPolicyMismatch });
    return baseline;
  } catch (error) {
    if (error?.code === "ENOENT" && !required) return null;
    if (error?.code === "ENOENT") {
      throw new Error(
        `Architecture baseline is missing at ${baselinePath}. Create it deliberately with --write-baseline --force-with-adr <reference>.`,
      );
    }
    throw error;
  }
};

const formatIssue = (issue) => {
  const change =
    issue.type === "new"
      ? `new count ${issue.currentCount}`
      : `${issue.baselineCount} -> ${issue.currentCount}`;
  return `  [${issue.category}] ${change}: ${issue.key}`;
};

const printSummary = (metrics) => {
  const summary = CATEGORY_ORDER.map(
    (category) =>
      `${category}=${metrics.entryCounts[category]} keys/${metrics.totals[category]} units`,
  ).join(", ");
  process.stdout.write(
    `Scanned ${metrics.productionFileCount} production TypeScript files (${summary}).\n`,
  );
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

  const { metrics, snapshot: current } = await scanArchitecture({ appRoot });
  printSummary(metrics);

  if (!options.writeBaseline) {
    const baseline = await readBaseline({ required: true });
    const issues = compareSnapshots(baseline, current);
    if (issues.length > 0) {
      process.stderr.write(
        `Architecture debt grew or moved in ${issues.length} place(s):\n${issues
          .map(formatIssue)
          .join("\n")}\n`,
      );
      process.exitCode = 1;
      return;
    }
    process.stdout.write("Architecture debt ratchet passed.\n");
    return;
  }

  const existing = await readBaseline({
    allowPolicyMismatch: Boolean(options.forceAdrReference),
    required: false,
  });
  const prepared = prepareSnapshotForWrite({
    current,
    existing,
    forceAdrReference: options.forceAdrReference,
  });

  if (!prepared.allowed) {
    const hardBlockMessage =
      prepared.hardBlockedIssues.length > 0
        ? "\n\nOversized-file debt is a hard ceiling. Reduce the file or split a cohesive behavior slice; --force-with-adr cannot approve this growth."
        : "\n\nUse --force-with-adr <reference> only after the growth is approved and documented.";
    process.stderr.write(
      `Refusing to bless ${prepared.issues.length} new, grown, or moved debt item(s):\n${prepared.issues
        .map(formatIssue)
        .join("\n")}${hardBlockMessage}\n`,
    );
    process.exitCode = 1;
    return;
  }

  await fs.writeFile(baselinePath, formatSnapshot(prepared.snapshot), "utf8");
  const forceMessage = options.forceAdrReference
    ? ` with approval reference ${options.forceAdrReference}`
    : " after debt reduction";
  process.stdout.write(`Updated ${baselinePath}${forceMessage}.\n`);
};

await main().catch((error) => {
  process.stderr.write(`${error.stack ?? error.message}\n`);
  process.exitCode = 1;
});
