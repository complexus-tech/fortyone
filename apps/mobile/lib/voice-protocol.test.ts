import assert from "node:assert/strict";
import test from "node:test";
import {
  nativeVoiceRoute,
  prepareVoiceTool,
  redactVoiceToolResult,
  voiceConversationContext,
  voiceSessionRequest,
  voiceTranscriptUpdate,
} from "./voice-protocol.ts";

test("voice session payload matches the deployed web contract without client metadata", () => {
  const context = {
    client: "mobile",
    currentPath: "/work/APP-12",
    workspace: "private-workspace",
    messages: [
      {
        id: "local-transcript",
        role: "user" as const,
        text: "What should I work on?",
        createdAt: 123,
      },
      { role: "assistant" as const, text: "Let me check your tasks." },
    ],
  };
  assert.deepEqual(voiceSessionRequest(context), {
    currentPath: "/work/APP-12",
    messages: [
      { role: "user", text: "What should I work on?" },
      { role: "assistant", text: "Let me check your tasks." },
    ],
  });
  assert.equal(context.client, "mobile");
  assert.equal(context.messages[0].id, "local-transcript");
});

test("voice session payload retains valid empty context and bounded transcript text", () => {
  assert.deepEqual(
    voiceSessionRequest({ currentPath: "/maya", messages: [] }),
    {
      currentPath: "/maya",
      messages: [],
    },
  );
  const messages = voiceConversationContext([
    { role: "user", text: "🦊".repeat(4001) },
  ]);
  assert.deepEqual(
    voiceSessionRequest({ currentPath: "/maya", messages }).messages,
    messages,
  );
  assert.equal(Array.from(messages[0].text).length, 4000);
});

test("provider arguments cannot approve a write or pass a signing token", () => {
  assert.deepEqual(
    prepareVoiceTool(
      "delete_story",
      JSON.stringify({
        reference: "APP-12",
        confirmed: true,
        confirmationToken: "invented",
      }),
    ),
    {
      mutation: true,
      args: { reference: "APP-12", confirmed: false },
    },
  );
  assert.equal(
    prepareVoiceTool("story_comments", '{"action":"add"}').mutation,
    true,
  );
  assert.equal(
    prepareVoiceTool("story_comments", '{"action":"list"}').mutation,
    false,
  );
  assert.equal(
    prepareVoiceTool("notifications", '{"action":"mark_all_read"}').mutation,
    true,
  );
  assert.throws(() => prepareVoiceTool("delete_workspace", "{}"));
  assert.throws(() => prepareVoiceTool("notifications", '{"action":"delete"}'));
  assert.throws(() => prepareVoiceTool("create_task", "[]"));
  assert.deepEqual(
    prepareVoiceTool(
      "create_task",
      '{"Confirmed":true,"ConfirmationToken":"invented"}',
    ).args,
    { confirmed: false },
  );
});

test("confirmation secrets and client commands never reach provider output", () => {
  assert.deepEqual(
    redactVoiceToolResult({
      success: false,
      confirmationToken: "secret",
      nested: [{ clientSecret: "secret", title: "Retain" }],
      clientAction: { path: "/settings" },
    }),
    { success: false, nested: [{ title: "Retain" }] },
  );
});

test("voice context respects server limits without splitting Unicode code points", () => {
  const messages = voiceConversationContext(
    Array.from({ length: 30 }, () => ({
      role: "user" as const,
      text: "🦊".repeat(4001),
    })),
  );
  assert.equal(messages.length, 24);
  assert.equal(Array.from(messages[0].text).length, 4000);
});

test("reconnecting deduplicates saved voice IDs without merging distinct equal messages", () => {
  assert.deepEqual(
    voiceConversationContext([
      { id: "voice-session-user", role: "user", text: "Hello" },
      { id: "voice-session-assistant", role: "assistant", text: "Hi" },
      { id: "voice-session-user", role: "user", text: "Hello" },
      { id: "voice-session-assistant", role: "assistant", text: "Hi" },
      { id: "another-turn", role: "user", text: "Hello" },
    ]),
    [
      { role: "user", text: "Hello" },
      { role: "assistant", text: "Hi" },
      { role: "user", text: "Hello" },
    ],
  );
});

test("native navigation uses an explicit route map and excludes billing/external URLs", () => {
  const id = "01234567-1234-4234-8234-0123456789ab";
  assert.equal(nativeVoiceRoute(`/work/${id}`), `/story/${id}`);
  assert.equal(nativeVoiceRoute(`/teams/${id}/stories`), `/teams/${id}`);
  assert.equal(nativeVoiceRoute("/notifications"), "/inbox");
  for (const path of [
    "//evil.test",
    "https://evil.test",
    "/settings/workspace/billing",
    "/settings?redirect=evil",
    "/work/APP-1",
    "/teams/%2F/stories",
  ])
    assert.equal(nativeVoiceRoute(path), null);
});

test("completed transcripts replace deltas for both modern and legacy event names", () => {
  assert.deepEqual(
    voiceTranscriptUpdate({
      type: "conversation.item.input_audio_transcription.completed",
      item_id: "item",
      transcript: "Create a task",
    }),
    { id: "item", text: "Create a task", role: "user", final: true },
  );
  assert.equal(
    voiceTranscriptUpdate({
      type: "response.audio_transcript.delta",
      item_id: "a",
      delta: "Hello",
    })?.role,
    "assistant",
  );
  assert.equal(
    voiceTranscriptUpdate({
      type: "response.output_audio_transcript.done",
      item_id: "a",
      transcript: "Hello",
    })?.final,
    true,
  );
  assert.equal(voiceTranscriptUpdate({ type: "session.created" }), null);
});
