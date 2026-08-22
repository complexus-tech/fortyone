import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import {
  getAgentMarkdown,
  getAgentNotFoundMarkdown,
  llmsText,
} from "./agent-content.ts";

void test("homepage Markdown is substantive and provides recovery links", () => {
  const markdown = getAgentMarkdown("/");
  assert.ok(markdown);
  assert.ok(markdown.length > 500);
  assert.match(markdown, /^# FortyOne/m);
  assert.match(markdown, /Developer resources/);
  assert.match(markdown, /openapi\.json/);
});

void test("unknown paths receive a concise Markdown 404", () => {
  const markdown = getAgentNotFoundMarkdown("/missing");
  assert.match(markdown, /^# 404: Page not found/m);
  assert.match(markdown, /sitemap\.xml/);
  assert.match(markdown, /llms\.txt/);
});

void test("llms.txt includes when-to-use guidance without advertising an About page", () => {
  assert.match(llmsText, /## When to use FortyOne/);
  assert.match(llmsText, /OpenAPI description/);
  assert.doesNotMatch(llmsText, /\/about/);
});

void test("published OpenAPI operations are uniquely identified and documented", () => {
  const document = JSON.parse(
    readFileSync(new URL("../../public/openapi.json", import.meta.url), "utf8"),
  ) as {
    openapi: string;
    paths: Record<
      string,
      Record<string, { description?: string; operationId?: string }>
    >;
  };

  assert.equal(document.openapi, "3.1.1");
  const operationIds = new Set<string>();
  for (const operations of Object.values(document.paths)) {
    for (const operation of Object.values(operations)) {
      assert.ok(operation.description);
      assert.ok(operation.operationId);
      assert.equal(operationIds.has(operation.operationId), false);
      operationIds.add(operation.operationId);
    }
  }
});
