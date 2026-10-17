/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest environment pragma.
import { readFile } from "node:fs/promises";
import path from "node:path";
import { Script } from "node:vm";
import { loadBindings, minify, transform } from "next/dist/build/swc";

it("keeps the chat link parser valid after production minification", async () => {
  // next/jest transforms TypeScript without the production minification pass.
  const filename = path.join(__dirname, "chat-links.ts");
  const source = await readFile(filename, "utf8");
  await loadBindings();
  const compiled: { code: string } = await transform(source, {
    filename,
    jsc: { parser: { syntax: "typescript" }, target: "es2017" },
    module: { type: "commonjs" },
  });
  const minified = await minify(compiled.code, {
    compress: true,
    mangle: true,
    ecma: 2017,
  });

  expect(() => new Script(minified.code, { filename })).not.toThrow();
});
