import { spawn, spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { createRequire } from "node:module";
import { homedir } from "node:os";
import { delimiter, join } from "node:path";
import { fileURLToPath } from "node:url";
import {
  getDesktopUrl,
  getProductionEnvironment,
  getProductionPackageArguments,
  getTauriArguments,
} from "./config.mjs";

const DESKTOP_DIR = fileURLToPath(new URL("..", import.meta.url));
const RUST_DIR = join(DESKTOP_DIR, "src-tauri");
const require = createRequire(import.meta.url);

function getEnvironment() {
  const env = { ...process.env };
  const cargo = spawnSync("cargo", ["--version"], { env, stdio: "ignore" });

  if (cargo.error?.code === "ENOENT") {
    const cargoBin = join(homedir(), ".cargo", "bin");
    const cargoName = process.platform === "win32" ? "cargo.exe" : "cargo";

    if (existsSync(join(cargoBin, cargoName))) {
      env.PATH = `${cargoBin}${delimiter}${env.PATH ?? ""}`;
    }
  }

  return env;
}

function verifyPrerequisites(env) {
  for (const command of ["cargo", "rustc"]) {
    const result = spawnSync(command, ["--version"], {
      env,
      encoding: "utf8",
    });

    if (result.error || result.status !== 0) {
      throw new Error(
        `Rust tooling (${command}) is unavailable. Install stable Rust from https://rustup.rs, then retry. ${result.stderr?.trim() ?? ""}`,
      );
    }
  }

  if (process.platform === "darwin") {
    const result = spawnSync("xcrun", ["--find", "clang"], {
      env,
      encoding: "utf8",
    });

    if (result.error || result.status !== 0) {
      throw new Error(
        "Apple command line tools are unavailable. Run xcode-select --install, finish installation, then retry.",
      );
    }
  }
}

async function verifyWebApp(url) {
  let response;

  try {
    response = await fetch(url, {
      redirect: "manual",
      signal: AbortSignal.timeout(30_000),
    });
  } catch (error) {
    throw new Error(
      `Cannot reach the web app at ${url}. Start it with pnpm dev:projects in another terminal and wait until it is ready, then retry pnpm dev:desktop. For a different local port, set FORTYONE_DESKTOP_URL. ${error.cause?.message ?? error.message}`,
    );
  } finally {
    await response?.body?.cancel();
  }

  if (response.status >= 500) {
    throw new Error(
      `The web app at ${url} returned HTTP ${response.status}. Resolve the error in the web app terminal, then retry.`,
    );
  }
}

function run(command, args, env, cwd = DESKTOP_DIR) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, { cwd, env, stdio: "inherit" });

    child.once("error", reject);
    child.once("exit", (code, signal) => {
      if (signal) {
        process.exitCode = signal === "SIGINT" ? 130 : 1;
        resolve(false);
        return;
      }

      process.exitCode = code ?? 1;
      resolve(code === 0);
    });
  });
}

async function main() {
  const [command, ...args] = process.argv.slice(2);

  if (!["dev", "build", "package", "check", "tauri"].includes(command)) {
    throw new Error(
      "Use pnpm dev:desktop, pnpm check:desktop, pnpm build:desktop, or pnpm package:desktop.",
    );
  }

  if (command !== "tauri" && args.length) {
    throw new Error(
      "Desktop commands do not accept extra flags. Set FORTYONE_DESKTOP_URL for local dev/build only; package:desktop always uses the fixed production configuration. Use pnpm --filter desktop tauri for direct CLI access.",
    );
  }

  const env = getEnvironment();

  if (command === "check") {
    verifyPrerequisites(env);

    for (const [executable, commandArgs, cwd] of [
      [
        process.execPath,
        ["--test", "scripts/config.test.mjs", "scripts/bridge.test.mjs"],
        DESKTOP_DIR,
      ],
      ["cargo", ["fmt", "--check"], RUST_DIR],
      ["cargo", ["check", "--locked"], RUST_DIR],
      ["cargo", ["test", "--locked"], RUST_DIR],
    ]) {
      if (!(await run(executable, commandArgs, env, cwd))) return;
    }

    return;
  }

  let cli;

  try {
    cli = require.resolve("@tauri-apps/cli/tauri.js");
  } catch {
    throw new Error(
      "The Tauri CLI is not installed. Run pnpm install at the repository root, then retry.",
    );
  }

  if (command === "tauri") {
    await run(process.execPath, [cli, ...args], env);
    return;
  }

  if (command === "package") {
    const packageArgs = getProductionPackageArguments();
    const productionEnv = getProductionEnvironment(env);
    verifyPrerequisites(productionEnv);
    console.log("Packaging FortyOne for macOS at https://cloud.fortyone.app");
    await run(process.execPath, [cli, ...packageArgs], productionEnv);
    return;
  }

  const url = getDesktopUrl(env.FORTYONE_DESKTOP_URL);
  env.FORTYONE_DESKTOP_URL = url;
  verifyPrerequisites(env);

  if (command === "dev") {
    await verifyWebApp(url);
    console.log(`Opening FortyOne Desktop at ${url}`);
  }

  await run(process.execPath, [cli, ...getTauriArguments(command, url)], env);
}

main().catch((error) => {
  console.error(`\nFortyOne Desktop: ${error.message}`);
  process.exitCode = 1;
});
