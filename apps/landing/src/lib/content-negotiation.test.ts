import assert from "node:assert/strict";
import test from "node:test";
import { preferredRepresentation } from "./content-negotiation.ts";

void test("defaults to HTML when Accept is absent", () => {
  assert.equal(preferredRepresentation(null), "text/html");
});

void test("honors quality values and client order", () => {
  assert.equal(
    preferredRepresentation("text/markdown, text/html;q=0.8"),
    "text/markdown",
  );
  assert.equal(
    preferredRepresentation("text/html, text/markdown"),
    "text/html",
  );
});

void test("specific exclusions override wildcards", () => {
  assert.equal(
    preferredRepresentation("*/*;q=0.5, text/markdown;q=0"),
    "text/html",
  );
});

void test("returns null when neither representation is acceptable", () => {
  assert.equal(preferredRepresentation("application/json"), null);
});
