import { spawn, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import { existsSync } from "node:fs";
import {
  copyFile,
  mkdir,
  mkdtemp,
  readdir,
  readFile,
  writeFile,
} from "node:fs/promises";
import { createRequire } from "node:module";
import { cpus, tmpdir } from "node:os";
import { basename, dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const DIRECTORY = dirname(fileURLToPath(import.meta.url));
const ROOT = fileURLToPath(new URL("../../../../..", import.meta.url));
const SOURCE_PREFIX = "apps/projects/src/modules/strategy";
const HELP = `Opt-in macOS WKWebView Strategy Map motion comparison (~20 seconds).

node apps/desktop/scripts/performance/strategy-map/run.mjs --run [--baseline <git-ref>]

Compares the selected committed strategy modules (default HEAD) with the working
copy. Uses real canvas, viewport, geometry, query, and node modules, but synthetic
500-objective data and representative card content. No auth, network, or writes
to account data. Requires installed workspace dependencies, a completed Projects
build (for CSS), and Apple command-line tools. Keep the window visible/focused.
Avoid concurrent heavy builds. All generated files/results stay in a temp folder.
No run occurs without --run. This tool is not connected to package/build scripts.`;

const hash = (value) => createHash("sha256").update(value).digest("hex");
const output = (command, args) => {
  const result = spawnSync(command, args, { cwd: ROOT, encoding: "utf8" });
  if (result.status !== 0)
    throw new Error(result.stderr || `${command} failed.`);
  return result.stdout.trim();
};
const execute = (command, args) =>
  new Promise((resolve, reject) => {
    const child = spawn(command, args, { stdio: "inherit" });
    child.once("error", reject);
    child.once("exit", (code, signal) =>
      code === 0
        ? resolve()
        : reject(new Error(`${command} exited with ${signal ?? code}.`)),
    );
  });

async function buildFixture({ directory, mode, baseline, esbuild, css }) {
  const fixtureDirectory = join(directory, mode);
  const sourceDirectory = join(fixtureDirectory, "strategy");
  await mkdir(sourceDirectory, { recursive: true });
  const files =
    mode === "baseline"
      ? output("git", [
          "ls-tree",
          "-r",
          "--name-only",
          baseline,
          SOURCE_PREFIX,
        ]).split("\n")
      : (await readdir(join(ROOT, SOURCE_PREFIX))).map(
          (name) => `${SOURCE_PREFIX}/${name}`,
        );
  const sourceHashes = {};
  for (const file of files) {
    if (!/\.(ts|tsx)$/.test(file) || /\.test\./.test(file)) continue;
    const source =
      mode === "baseline"
        ? output("git", ["show", `${baseline}:${file}`])
        : await readFile(join(ROOT, file), "utf8");
    await writeFile(join(sourceDirectory, basename(file)), source);
    sourceHashes[file] = hash(source);
  }
  const mocks = {
    "@/hooks": `export const useWorkspacePath=()=>({workspaceSlug:"motion-fixture",withWorkspace:(path)=>"/motion-fixture"+path});`,
    "@/lib/auth/client": `const session={token:"synthetic"};export const useSession=()=>({data:session});`,
    "@/lib/hooks/members": `const data=[];export const useMembers=()=>({data});`,
    "@/lib/hooks/objective-statuses": `const data=[];export const useObjectiveStatuses=()=>({data});`,
    "@/modules/objectives/hooks": `const mutation={mutate:()=>{}};export const useUpdateObjectiveMutation=()=>mutation;`,
    "@/modules/objectives/stories/overview/new-key-result": `export const NewKeyResultDialog=()=>null;`,
    "@/modules/teams/public/client": `const data=[];export const useTeams=()=>({data});`,
    "@/modules/key-results/components/key-result-context-menu": `export const KeyResultContextMenu=({children})=>children;`,
    "@/modules/objectives/queries/get-key-results": `export const getKeyResults=async(id)=>window.fixtureKeyResults.get(id)||[];`,
    ui: `import {forwardRef} from "react";
      export const Box=forwardRef(({children,...props},ref)=><div {...props} ref={ref}>{children}</div>);
      export const Flex=forwardRef(({children,align,justify,...props},ref)=><div {...props} ref={ref} style={{display:"flex",alignItems:align==="center"?"center":undefined,justifyContent:justify,...props.style}}>{children}</div>);
      export const Text=({children,color,fontWeight,...props})=><span {...props}>{children}</span>;`,
    lib: `export const cn=(...args)=>args.flat(Infinity).filter(value=>typeof value==="string"&&value).join(" ");`,
  };
  await esbuild.build({
    entryPoints: [join(directory, "fixture.tsx")],
    outfile: join(fixtureDirectory, "fixture.js"),
    bundle: true,
    platform: "browser",
    format: "iife",
    jsx: "automatic",
    define: {
      "process.env.NODE_ENV": '"production"',
      FIXTURE_MODE: JSON.stringify(mode),
    },
    nodePaths: [
      join(ROOT, "apps/projects/node_modules"),
      join(ROOT, "node_modules"),
    ],
    plugins: [
      {
        name: "isolated-strategy-fixture",
        setup(build) {
          build.onResolve({ filter: /.*/ }, (args) => {
            if (Object.hasOwn(mocks, args.path))
              return { path: args.path, namespace: "fixture-mock" };
            if (args.path === "./strategy-map-cards")
              return { path: join(directory, "card-fixture.tsx") };
            if (args.path === "@strategy-fixture/canvas")
              return { path: join(sourceDirectory, "strategy-map-canvas.tsx") };
            if (args.path === "@strategy-fixture/layout")
              return { path: join(sourceDirectory, "strategy-map-layout.ts") };
            if (args.path.startsWith("@/")) {
              const base = join(ROOT, "apps/projects/src", args.path.slice(2));
              const resolved = [".ts", ".tsx", "/index.ts", "/index.tsx"]
                .map((suffix) => `${base}${suffix}`)
                .find(existsSync);
              if (resolved) return { path: resolved };
            }
            return undefined;
          });
          build.onLoad({ filter: /.*/, namespace: "fixture-mock" }, (args) => ({
            contents: mocks[args.path],
            loader: "tsx",
            resolveDir: join(ROOT, "apps/projects"),
          }));
        },
      },
    ],
    logLevel: "warning",
  });
  await writeFile(
    join(fixtureDirectory, "index.html"),
    `<!doctype html><html><head><meta charset="utf-8"><style>${css}</style></head><body><div id="root"></div><script src="fixture.js"></script></body></html>`,
  );
  return {
    sourceHashes,
    bundleHash: hash(await readFile(join(fixtureDirectory, "fixture.js"))),
  };
}

async function run(baselineRef) {
  if (process.platform !== "darwin")
    throw new Error("This fixture requires macOS WKWebView.");
  const baseline = output("git", [
    "rev-parse",
    "--verify",
    `${baselineRef}^{commit}`,
  ]);
  output("xcrun", ["--find", "swiftc"]);
  // esbuild is already installed for the docs workspace; do not install or
  // modify dependencies simply to run a local opt-in diagnostic.
  const docsRequire = createRequire(join(ROOT, "apps/docs/package.json"));
  const esbuild = createRequire(docsRequire.resolve("fumadocs-mdx"))("esbuild");
  const chunksDirectory = join(ROOT, "apps/projects/.next/static/chunks");
  const cssFiles = (await readdir(chunksDirectory))
    .filter((file) => file.endsWith(".css"))
    .sort();
  if (!cssFiles.length)
    throw new Error(
      "Build Projects once before running: its generated CSS is required.",
    );
  const css =
    (
      await Promise.all(
        cssFiles.map((file) => readFile(join(chunksDirectory, file), "utf8")),
      )
    ).join("\n") + (await readFile(join(DIRECTORY, "fixture.css"), "utf8"));
  const directory = await mkdtemp(join(tmpdir(), "fortyone-strategy-motion-"));
  console.log(`Artifacts: ${directory}`);
  for (const file of ["fixture.tsx", "card-fixture.tsx"])
    await copyFile(join(DIRECTORY, file), join(directory, file));
  const metadata = {
    startedAt: new Date().toISOString(),
    baseline,
    esbuild: esbuild.version,
    os: output("sw_vers", ["-productVersion"]),
    osBuild: output("sw_vers", ["-buildVersion"]),
    hardware: output("sysctl", ["-n", "hw.model"]),
    cpu: cpus()[0]?.model,
    viewport: [1440, 850],
    cssHash: hash(css),
    fixtureHash: hash(await readFile(join(DIRECTORY, "fixture.tsx"))),
    arms: {},
  };
  // Freeze both arms before launching either native window.
  for (const mode of ["baseline", "after"])
    metadata.arms[mode] = await buildFixture({
      directory,
      mode,
      baseline,
      esbuild,
      css,
    });
  const executable = join(directory, "harness");
  await execute("xcrun", [
    "swiftc",
    join(DIRECTORY, "Harness.swift"),
    "-o",
    executable,
  ]);
  const results = { metadata, runs: {} };
  for (const mode of ["baseline", "after"]) {
    await execute(executable, [join(directory, mode)]);
    results.runs[mode] = JSON.parse(
      await readFile(join(directory, mode, "results.json"), "utf8"),
    );
  }
  metadata.endedAt = new Date().toISOString();
  await writeFile(
    join(directory, "comparison.json"),
    JSON.stringify(results, null, 2),
  );
  const valid = Object.values(results.runs).every(
    (events) =>
      events.some((event) => event.type === "complete") &&
      !events.some(
        (event) => event.type === "fatal" || event.type === "timeout",
      ) &&
      events
        .filter((event) => ["ready", "phase", "complete"].includes(event.type))
        .every((event) => event.focused && event.visible === "visible"),
  );
  console.log(`Comparison saved: ${join(directory, "comparison.json")}`);
  if (!valid)
    throw new Error(
      "A fixture failed or lost focus; inspect the preserved events before interpreting measurements.",
    );
}

const args = process.argv.slice(2);
if (args.length === 0 || args[0] === "--help") console.log(HELP);
else if (
  args[0] === "--run" &&
  (args.length === 1 || (args.length === 3 && args[1] === "--baseline"))
) {
  await run(args[2] ?? "HEAD").catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
  });
} else {
  console.error(HELP);
  process.exitCode = 1;
}
