import { readdir } from "node:fs/promises";
import { spawnSync } from "node:child_process";

const ignored = new Set([
  "node_modules",
  ".expo",
  "ios",
  "android",
  "dist",
  ".git",
]);
async function testFiles(directory) {
  const files = [];
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    if (ignored.has(entry.name)) continue;
    const path = `${directory}/${entry.name}`;
    if (entry.isDirectory()) files.push(...(await testFiles(path)));
    else if (entry.name.endsWith(".test.ts")) files.push(path);
  }
  return files;
}

const files = (await testFiles(".")).sort();
if (!files.length) throw new Error("No mobile tests were found.");
const result = spawnSync(
  process.execPath,
  ["--import", "tsx", "--test", ...files],
  { stdio: "inherit" },
);
if (result.error) throw result.error;
process.exit(result.status ?? 1);
