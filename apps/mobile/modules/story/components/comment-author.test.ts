import assert from "node:assert/strict";
import test from "node:test";
import { FORMER_USER_ID } from "@/lib/former-user";
import { resolveCommentAuthor } from "./comment-author.ts";

test("retained comments identify Former user without any active membership or stale avatar", () => {
  assert.deepEqual(resolveCommentAuthor({ userId: FORMER_USER_ID }, []), {
    name: "Former user",
    avatarUrl: undefined,
  });
  assert.deepEqual(
    resolveCommentAuthor(
      {
        userId: FORMER_USER_ID,
        user: {
          id: FORMER_USER_ID,
          fullName: "Stale name",
          username: "stale",
          avatarUrl: "https://example.com/stale.png",
          isSystem: true,
        },
      },
      [],
    ),
    { name: "Former user", avatarUrl: undefined },
  );
});

test("departed and system comment authors use matching embedded identity without inventing former users", () => {
  const user = {
    id: "maya",
    fullName: "Maya",
    username: "maya",
    isSystem: true,
  };
  assert.equal(resolveCommentAuthor({ userId: "maya", user }, []).name, "maya");
  assert.equal(
    resolveCommentAuthor({ userId: "missing", user }, []).name,
    "Unknown",
  );
});
