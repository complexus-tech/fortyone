import assert from "node:assert/strict";
import test from "node:test";
import {
  messageLinkURL,
  resultStories,
  toolResultError,
} from "./message-model.ts";

const first = "01234567-1234-4234-8234-0123456789ab";
const second = "01234567-1234-4234-8234-0123456789ac";

test("grouped story tool results retain source order, references and status IDs", () => {
  // listTeamStories returns groups inside stories, not a flat result list.
  const result = resultStories({
    success: true,
    kind: "story-list",
    stories: [
      {
        key: "started",
        stories: [
          {
            id: first,
            title: "First task",
            sequenceId: 42,
            team: { code: "ENG" },
            statusId: "status-started",
          },
        ],
      },
      { key: "backlog", stories: [{ id: second, title: "Second task" }] },
      { key: "repeated", stories: [{ id: first, title: "Duplicate" }] },
    ],
  });
  assert.deepEqual(
    result.map((story) => story.id),
    [first, second],
  );
  assert.equal(result[0].reference, "ENG-42");
  assert.equal(result[0].statusId, "status-started");
  assert.equal(result[0].title, "First task");
});

test("search supports both enriched and flattened status metadata without treating objectives as tasks", () => {
  const result = resultStories({
    stories: [
      {
        id: first,
        title: "Enriched search result",
        status: { name: "In review", category: "started", color: "#f43f5e" },
      },
      {
        id: second,
        title: "Flat search result",
        statusName: "Finished",
        statusCategory: "completed",
        statusColor: "#0f0",
        teamCode: "WEB",
        sequenceId: 12,
      },
    ],
    objectives: [{ id: first, name: "Objective" }],
  });
  assert.deepEqual(result[0].status, {
    name: "In review",
    category: "started",
    color: "#f43f5e",
  });
  assert.deepEqual(result[1].status, {
    name: "Finished",
    category: "completed",
    color: "#0f0",
  });
  assert.equal(result[1].reference, "WEB-12");
});

test("creation summaries and partially successful batches remain navigable", () => {
  assert.equal(
    resultStories({ success: true, story: { id: first, title: "Created" } })[0]
      .id,
    first,
  );
  const result = resultStories({
    success: false,
    stories: [{ id: first, title: "Created" }],
    failedStories: [{ title: "Could not create", error: "Team unavailable" }],
  });
  assert.deepEqual(
    result.map((story) => story.title),
    ["Created"],
  );
});

test("generic non-story entities, malformed IDs and nested untrusted payloads are not story links", () => {
  for (const output of [
    { data: [{ id: first, title: "Document" }] },
    { results: [{ id: first, title: "Feedback" }] },
    { objectives: [{ id: first, title: "Objective" }] },
    { stories: [{ id: "https://example.com", title: "External target" }] },
    { stories: [{ id: first, title: " " }] },
    { stories: [{ stories: [{ data: { id: first, title: "Nested" } }] }] },
    null,
  ])
    assert.deepEqual(resultStories(output), []);
});

test("unknown status categories and malformed native colors cannot reach status icons", () => {
  const result = resultStories({
    stories: [
      { id: first, title: "Unknown", status: { category: "invented" } },
      {
        id: second,
        title: "Bad color",
        status: { category: "started", color: "url(https://example.com)" },
      },
    ],
  });
  assert.equal(result[0].status, undefined);
  assert.equal(result[1].status?.category, "started");
  assert.equal(result[1].status?.color, undefined);
});

test("message links only open valid HTTP pages and reject unsupported or credential-bearing URLs", () => {
  assert.equal(
    messageLinkURL("https://example.com/help?q=maya"),
    "https://example.com/help?q=maya",
  );
  for (const target of [
    "javascript:alert(1)",
    "data:text/html,test",
    "file:///private/file",
    "/settings",
    "fortyone://settings",
    "https://user:password@example.com",
    "http://",
  ])
    assert.equal(messageLinkURL(target), null);
});

test("tool summaries stay hidden while failures remain actionable", () => {
  for (const message of [
    "Found 0 stories in this team.",
    "Found 1 story in this team.",
    "Task updated.",
  ]) {
    assert.equal(toolResultError({ success: true, message }), null);
  }
  assert.equal(
    toolResultError({ success: false, message: "Access denied" }),
    "Access denied",
  );
  assert.equal(
    toolResultError({ error: "Network unavailable" }),
    "Network unavailable",
  );
  assert.equal(
    toolResultError({ success: false }),
    "This action could not be completed.",
  );
});
