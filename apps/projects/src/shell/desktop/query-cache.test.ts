/* global expect, it -- Jest globals. */
import { TextEncoder } from "node:util";
import { QueryClient } from "@tanstack/react-query";
import { createDesktopCache, restoreDesktopCache } from "./query-cache";

Object.assign(globalThis, { TextEncoder });
const scope = {
  accountId: "user-a",
  workspaceId: "workspace-a",
  workspaceSlug: "acme",
};
const now = Date.now();

it("persists domain data without authentication, integration secrets or another workspace", () => {
  const client = new QueryClient();
  client.setQueryData(["stories", "acme", "list"], [{ id: "cached-story" }]);
  client.setQueryData(["users", "profile"], {
    id: "user-a",
    fullName: "Account A",
  });
  client.setQueryData(
    ["stories", "another", "list"],
    [{ id: "private-story" }],
  );
  client.setQueryData(["session"], { token: "never-persist" });
  client.setQueryData(
    ["developer-settings", "acme", "oauth-applications"],
    [{ secret: "never-persist" }],
  );
  const saved = createDesktopCache(client, scope);
  expect(saved.queries).toHaveLength(2);
  expect(JSON.stringify(saved)).not.toContain("never-persist");
  expect(JSON.stringify(saved)).not.toContain("private-story");
});
it("restores cached views but preserves newer server hydration", () => {
  const client = new QueryClient();
  const key = ["stories", "acme", "list"];
  client.setQueryData(key, "server", { updatedAt: now });
  const queries = [
    { queryKey: key, data: "older-cache", updatedAt: now - 1000 },
    {
      queryKey: ["workspaces", "settings", "acme"],
      data: { storyTerm: "task" },
      updatedAt: now - 500,
    },
  ];
  expect(
    restoreDesktopCache(
      client,
      scope,
      { ...scope, version: 1, savedAt: now, queries },
      now,
    ),
  ).toBe(1);
  expect(client.getQueryData(key)).toBe("server");
  expect(client.getQueryData(["workspaces", "settings", "acme"])).toEqual({
    storyTerm: "task",
  });
});
it("rejects other accounts, expired data and an unrelated query embedded in a valid envelope", () => {
  const client = new QueryClient();
  const envelope = {
    ...scope,
    version: 1,
    savedAt: now,
    queries: [
      {
        queryKey: ["stories", "acme", "list"],
        data: "private",
        updatedAt: now,
      },
    ],
  };
  expect(
    restoreDesktopCache(
      client,
      { ...scope, accountId: "user-b" },
      envelope,
      now,
    ),
  ).toBe(0);
  expect(restoreDesktopCache(client, scope, envelope, now + 8 * 86400000)).toBe(
    0,
  );
  expect(
    restoreDesktopCache(
      client,
      scope,
      {
        ...envelope,
        queries: [
          {
            queryKey: ["stories", "other", "list"],
            data: "private",
            updatedAt: now,
          },
        ],
      },
      now,
    ),
  ).toBe(0);
  expect(client.getQueryCache().getAll()).toHaveLength(0);
});
it("does not persist a stale unscoped profile belonging to another account", () => {
  const client = new QueryClient();
  client.setQueryData(["users", "profile"], { id: "user-b" });
  expect(createDesktopCache(client, scope).queries).toHaveLength(0);
});

it("omits temporary upload rows and never serializes webview blob URLs", () => {
  const client = new QueryClient();
  const attachments = [
    { id: "temp-upload", filename: "photo.jpg", url: "blob:temporary" },
    { id: "durable-upload", filename: "photo.jpg", url: "", pending: true },
    {
      id: "server-upload",
      filename: "saved.jpg",
      url: "https://files.test/saved",
    },
  ];
  const key = ["stories", "acme", "detail", "story", "attachments"];
  client.setQueryData(key, attachments);
  client.setQueryData(["documents", "acme", "detail", "doc"], {
    contentHtml: '<img src="blob:temporary-media">',
  });
  const saved = createDesktopCache(client, scope);
  expect(saved.queries).toHaveLength(1);
  expect(saved.queries[0].data).toEqual(attachments.slice(1));
  expect(JSON.stringify(saved)).not.toContain("blob:");
  expect(client.getQueryData(key)).toEqual(attachments);
  expect(
    restoreDesktopCache(new QueryClient(), scope, {
      ...saved,
      queries: [{ queryKey: key, data: attachments, updatedAt: now }],
    }),
  ).toBe(0);
});

it("retains custom-field definitions only in their owning workspace", () => {
  const client = new QueryClient();
  client.setQueryData(
    ["custom-fields", "acme", "team", "team"],
    [{ id: "required-field", required: true }],
  );
  client.setQueryData(
    ["custom-fields", "elsewhere", "team", "private"],
    [{ id: "private-field" }],
  );
  const cached = createDesktopCache(client, scope);
  expect(cached.queries).toHaveLength(1);
  expect(JSON.stringify(cached)).toContain("required-field");
  expect(JSON.stringify(cached)).not.toContain("private-field");
});
