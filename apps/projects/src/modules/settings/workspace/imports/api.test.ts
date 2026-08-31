/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { ImportStoriesRequest } from "./api";
import { buildImportStoryRequests } from "./api";

const story = (description: string) => ({
  description,
  priority: "No Priority" as const,
  teamId: "team-1",
  title: "Imported task",
});

const requestFor = (items: ImportStoriesRequest["items"]) => ({
  items,
  provider: "file" as const,
  sourceDigest: "a".repeat(64),
});

describe("import API batching", () => {
  it("keeps every request within the server's 50-item contract", () => {
    const requests = buildImportStoryRequests(
      requestFor(
        Array.from({ length: 101 }, (_, index) => ({
          sourceKey: `row-${index + 1}`,
          story: story("Short description"),
        })),
      ),
    );

    expect(requests.map(({ items }) => items.length)).toEqual([50, 50, 1]);
  });

  it("splits heavily escaped descriptions below the request body limit", () => {
    const requests = buildImportStoryRequests(
      requestFor(
        Array.from({ length: 20 }, (_, index) => ({
          sourceKey: `row-${index + 1}`,
          story: story("\u0000".repeat(20_000)),
        })),
      ),
    );

    expect(requests.length).toBeGreaterThan(1);
    for (const request of requests) {
      expect(
        new TextEncoder().encode(JSON.stringify(request)).byteLength,
      ).toBeLessThanOrEqual(850 * 1024);
    }
    expect(requests.flatMap(({ items }) => items)).toHaveLength(20);
  });
});
