import { spawn, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import { copyFile, mkdtemp, readFile, writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { cpus, freemem, release, tmpdir, totalmem } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { analyzeRun } from "./analyze.mjs";

const DIRECTORY = dirname(fileURLToPath(import.meta.url));
const ROOT = fileURLToPath(new URL("../../../..", import.meta.url));
const HELP = `Opt-in visible macOS WKWebView benchmark (~6 minutes).

Usage: node apps/desktop/scripts/performance/run.mjs --run

Requires installed workspace dependencies and Apple command-line tools.
Opens an isolated local fixture; no account, network, or production data.
Leave the benchmark visible and focused. Avoid builds during measurement.
All generated assets, binary, logs, and results go to a fresh temp directory.
No benchmark runs without --run; it is not part of any package/build script.`;

function output(command, args) {
  const result = spawnSync(command, args, { encoding: "utf8" });
  return result.status === 0 ? result.stdout.trim() : null;
}

function run(command, args) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, { stdio: "inherit" });
    child.once("error", reject);
    child.once("exit", (code, signal) => {
      if (code === 0) resolve();
      else reject(new Error(`${command} exited with ${signal ?? code}.`));
    });
  });
}

async function benchmark() {
  if (process.platform !== "darwin")
    throw new Error("This benchmark requires macOS WKWebView.");
  const swiftCompiler = output("xcrun", ["--find", "swiftc"]);
  if (!swiftCompiler) {
    throw new Error(
      "Swift is unavailable. Install Apple command-line tools with xcode-select --install.",
    );
  }
  const require = createRequire(join(ROOT, "packages/ui/package.json"));
  const modulePath = join(dirname(require.resolve("hyperellipse")), "index.js");
  const cssPath = join(ROOT, "packages/tailwind-config/squircle-fallback.css");
  const [moduleSource, css, planSource] = await Promise.all([
    readFile(modulePath, "utf8"),
    readFile(cssPath, "utf8"),
    readFile(join(DIRECTORY, "plan.json"), "utf8"),
  ]);
  const exported = "export { registerHyperellipse };";
  if (moduleSource.split(exported).length !== 2) {
    throw new Error(
      "Hyperellipse's module format changed; review the fixture adapter before measuring.",
    );
  }
  const directory = await mkdtemp(join(tmpdir(), "fortyone-squircle-perf-"));
  console.log(`Artifacts: ${directory}`);
  for (const name of ["fixture.js", "fixture.css", "plan.json"]) {
    await copyFile(join(DIRECTORY, name), join(directory, name));
  }
  await writeFile(
    join(directory, "hyperellipse-browser.js"),
    moduleSource.replace(
      exported,
      "window.registerHyperellipse = registerHyperellipse;",
    ),
  );
  await writeFile(join(directory, "fallback.css"), css);
  const hash = (value) => createHash("sha256").update(value).digest("hex");
  const metadata = {
    os: output("sw_vers", ["-productVersion"]),
    osBuild: output("sw_vers", ["-buildVersion"]),
    kernel: release(),
    hardware: output("sysctl", ["-n", "hw.model"]),
    cpu: cpus()[0]?.model,
    logicalCpus: cpus().length,
    totalMemoryBytes: totalmem(),
    freeMemoryBytes: freemem(),
    viewport: [1440, 850],
    sourceHashes: { hyperellipse: hash(moduleSource), sharedCSS: hash(css) },
  };
  await writeFile(
    join(directory, "metadata.json"),
    JSON.stringify(metadata, null, 2),
  );
  await run(swiftCompiler, [
    join(DIRECTORY, "Benchmark.swift"),
    "-o",
    join(directory, "benchmark"),
  ]);
  metadata.startedAt = new Date().toISOString();
  try {
    await run(join(directory, "benchmark"), [directory]);
  } finally {
    metadata.endedAt = new Date().toISOString();
    Object.assign(
      metadata,
      JSON.parse(
        await readFile(join(directory, "native-metadata.json"), "utf8"),
      ),
    );
    const events = JSON.parse(
      await readFile(join(directory, "events.json"), "utf8"),
    );
    const measured = {
      schemaVersion: 1,
      cpuEncoding: "milliseconds",
      metadata,
      plan: JSON.parse(planSource),
      events,
    };
    await writeFile(
      join(directory, "results.json"),
      JSON.stringify(measured, null, 2),
    );
    const summary = analyzeRun(measured);
    await writeFile(
      join(directory, "summary.json"),
      JSON.stringify(summary, null, 2),
    );
    console.log(`Results and matched-pair summary saved in ${directory}`);
    if (
      events.filter((event) => event.type === "trial-end").length !==
      measured.plan.length
    ) {
      throw new Error(
        "Measurement ended before every trial completed; partial results were preserved.",
      );
    }
    if (!summary.acceptedPairs.length) {
      throw new Error(
        "No comparable visible/focused pairs were collected; inspect the saved exclusions.",
      );
    }
  }
}

if (process.argv.length === 3 && process.argv[2] === "--run") {
  await benchmark().catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
  });
} else {
  console.log(HELP);
  if (process.argv.length > 2 && process.argv[2] !== "--help")
    process.exitCode = 1;
}
