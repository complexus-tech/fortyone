import { cp, mkdir, readFile } from "node:fs/promises";
import { createRequire } from "node:module";
import path from "node:path";
import { fileURLToPath } from "node:url";

const require = createRequire(import.meta.url);
const packagePath = require.resolve("pdfjs-dist/package.json");
const { version } = JSON.parse(await readFile(packagePath, "utf8"));
if (!/^\d+\.\d+\.\d+$/.test(version)) {
  throw new Error("PDF.js assets require a pinned release version.");
}
const packageRoot = path.dirname(packagePath);
const appRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const target = path.join(appRoot, "public", "pdfjs", version);
await mkdir(target, { recursive: true });
await Promise.all([
  cp(
    path.join(packageRoot, "legacy/build/pdf.worker.min.mjs"),
    path.join(target, "pdf.worker.min.mjs"),
  ),
  ...["cmaps", "standard_fonts", "wasm", "LICENSE"].map((name) =>
    cp(path.join(packageRoot, name), path.join(target, name), {
      recursive: true,
    }),
  ),
]);
