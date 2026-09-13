import assert from "node:assert/strict";
import test from "node:test";
import { parseNotificationsPage } from "./response.ts";

test("inbox contract accepts the server pagination envelope and rejects a legacy array", () => {
  const page = {
    notifications: [{ id: "notification" }],
    pagination: { page: 1, pageSize: 25, hasMore: true, nextPage: 2 },
  };
  assert.equal(parseNotificationsPage(page), page);
  assert.throws(
    () => parseNotificationsPage([{ id: "notification" }]),
    /inbox response/,
  );
  assert.throws(
    () =>
      parseNotificationsPage({
        ...page,
        pagination: { ...page.pagination, nextPage: 1 },
      }),
    /inbox response/,
  );
  assert.throws(() => parseNotificationsPage(null), /inbox response/);
});
