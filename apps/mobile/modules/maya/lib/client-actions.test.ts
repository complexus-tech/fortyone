import assert from "node:assert/strict";
import test from "node:test";
import type { MayaMessage } from "../types";
import {
  getMayaClientActions,
  hasMayaMutationReceipt,
} from "./client-actions.ts";
import {
  parseMayaSessions,
  prepareMobileMayaMessages,
} from "./chat-protocol.ts";

const teamId = "5a3a60c2-31bb-4e57-bf43-f9b213ab6743";
const entityId = "9ad0a1bc-dccc-4e9d-a523-740f95c80626";
const navigation = (
  input: unknown,
  output: unknown = { route: "https://attacker.test" },
): MayaMessage => ({
  id: "response",
  role: "assistant",
  parts: [
    {
      type: "tool-navigation",
      state: "output-available",
      toolCallId: "call",
      input,
      output,
    },
  ],
});

test("native navigation only handles validated supported semantic destinations", () => {
  assert.deepEqual(
    getMayaClientActions(navigation({ targetType: "billing" })),
    [],
  );
  assert.deepEqual(
    getMayaClientActions(
      navigation({ targetType: "story", entityId: "../../settings" }),
    ),
    [],
  );
  assert.deepEqual(
    getMayaClientActions(navigation({ targetType: "story", entityId })),
    [
      {
        type: "navigate",
        href: { pathname: "/story/[storyId]", params: { storyId: entityId } },
      },
    ],
  );
  assert.deepEqual(
    getMayaClientActions(
      navigation({ targetType: "objective", entityId, teamId }),
    ),
    [
      {
        type: "navigate",
        href: {
          pathname: "/teams/[teamId]",
          params: { teamId, objectiveId: entityId },
        },
      },
    ],
  );
  assert.deepEqual(
    getMayaClientActions(
      navigation({ targetType: "story", entityId }, { error: "Not found" }),
    ),
    [],
  );
});

test("approved terminal mutation receipts invalidate data, ordinary reads do not", () => {
  assert.equal(
    hasMayaMutationReceipt(navigation({ targetType: "summary" })),
    false,
  );
  const receipt: MayaMessage = {
    id: "result",
    role: "assistant",
    parts: [
      {
        type: "tool-deleteStory",
        state: "output-available",
        toolCallId: "call",
        input: {},
        output: { success: true },
        approval: { id: "approval", approved: true },
      },
    ],
  };
  assert.equal(hasMayaMutationReceipt(receipt), true);
});

test("mobile history keeps approval receipts and metadata while omitting binary attachments", () => {
  const message: MayaMessage = {
    id: "message",
    role: "user",
    metadata: { source: "voice" },
    parts: [
      {
        type: "file",
        url: "data:image/png;base64,binary",
        mediaType: "image/png",
      },
    ],
  };
  const result = prepareMobileMayaMessages([message]);
  assert.deepEqual(result[0].metadata, { source: "voice" });
  assert.equal(result[0].id, message.id);
  assert.equal(result[0].parts[0].type, "text");
  assert.equal(message.parts[0].type, "file");
});

test("conversation history rejects an unexpected account rather than exposing it", () => {
  assert.throws(
    () =>
      parseMayaSessions(
        [
          {
            id: "abcdefghijklmnop",
            userId: "other",
            workspaceId: teamId,
            title: "Private",
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        ],
        "user",
      ),
    /invalid conversation history/,
  );
});
