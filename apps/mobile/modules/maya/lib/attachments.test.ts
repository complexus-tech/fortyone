import assert from "node:assert/strict";
import test from "node:test";
import type { FileUIPart } from "ai";
import {
  createMayaAttachmentQueue,
  getMayaAttachmentMediaType,
  MAYA_ATTACHMENT_LIMITS,
  validateMayaAttachmentSizes,
  validateMayaFileParts,
  type MayaAttachment,
} from "./attachments.ts";

const file = (id = "one", size = 4): MayaAttachment => ({
  id,
  uri: `file:///cache/${id}`,
  name: `${id}.pdf`,
  mediaType: "application/pdf",
  size,
});
const part = (size: number, mediaType = "application/pdf"): FileUIPart => ({
  type: "file",
  mediaType,
  filename: "document.pdf",
  url: `data:${mediaType};base64,${Buffer.alloc(size).toString("base64")}`,
});

test("attachment limits accept exact boundaries and reject per-file, aggregate and count overflow", () => {
  const { fileBytes, totalBytes } = MAYA_ATTACHMENT_LIMITS;
  validateMayaAttachmentSizes([
    { size: fileBytes },
    { size: totalBytes - fileBytes },
  ]);
  assert.throws(
    () => validateMayaAttachmentSizes([{ size: fileBytes + 1 }]),
    /5 MB/,
  );
  assert.throws(
    () =>
      validateMayaAttachmentSizes([{ size: fileBytes }, { size: fileBytes }]),
    /8 MB/,
  );
  assert.throws(
    () =>
      validateMayaAttachmentSizes(
        Array.from({ length: 6 }, () => ({ size: 1 })),
      ),
    /5 images/,
  );
  for (const size of [0, -1, NaN, Infinity, 1.5])
    assert.throws(() => validateMayaAttachmentSizes([{ size }]), /size|empty/);
});

test("file parts allow only bounded matching MIME/base64 data URLs", () => {
  validateMayaFileParts([part(1), part(2, "image/png"), part(3, "image/jpeg")]);
  for (const invalid of [
    { ...part(4), mediaType: "image/svg+xml" },
    { ...part(4), mediaType: "image/png" },
    { ...part(4), url: "https://example.test/file.pdf" },
    { ...part(4), url: "file:///private/file.pdf" },
    { ...part(4), url: "data:application/pdf;base64,!!!!" },
    { ...part(4), url: "data:application/pdf;base64,A===" },
    { ...part(4), url: "data:application/pdf;base64," },
  ])
    assert.throws(() => validateMayaFileParts([invalid]));
  assert.throws(
    () => validateMayaFileParts([part(MAYA_ATTACHMENT_LIMITS.fileBytes + 1)]),
    /5 MB/,
  );
  assert.throws(
    () => validateMayaFileParts([part(5 * 1024 * 1024), part(4 * 1024 * 1024)]),
    /8 MB/,
  );
});

test("native header detection rejects a mislabeled SVG and accepts supported signatures", () => {
  for (const [bytes, mediaType] of [
    [[255, 216, 255], "image/jpeg"],
    [[137, 80, 78, 71, 13, 10, 26, 10], "image/png"],
    [[71, 73, 70, 56, 57, 97], "image/gif"],
    [[82, 73, 70, 70, 0, 0, 0, 0, 87, 69, 66, 80], "image/webp"],
    [[37, 80, 68, 70, 45], "application/pdf"],
  ] as const)
    assert.equal(getMayaAttachmentMediaType(Uint8Array.from(bytes)), mediaType);
  assert.throws(
    () => getMayaAttachmentMediaType(new TextEncoder().encode("<svg")),
    /Attach/,
  );
});

const fixture = (
  overrides: Partial<Parameters<typeof createMayaAttachmentQueue>[0]> = {},
) => {
  const released: string[] = [];
  const queue = createMayaAttachmentQueue({
    isCurrent: () => true,
    pick: async () => [file()],
    read: async (attachment) => ({ size: attachment.size, base64: "JVBERg==" }),
    release: (attachments) =>
      released.push(...attachments.map((item) => item.id)),
    ...overrides,
  });
  return { queue, released };
};

test("late picker completion after leaving a conversation cleans only newly owned copies", async () => {
  let finish!: (files: MayaAttachment[]) => void;
  const { queue, released } = fixture({
    pick: () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  });
  const pending = queue.pick();
  queue.blur();
  finish([file()]);
  await assert.rejects(pending, /conversation changed/);
  assert.deepEqual(queue.getSnapshot().attachments, []);
  assert.deepEqual(released, ["one"]);
});

test("picker-induced app deactivation preserves selection; ordinary backgrounding clears it", async () => {
  let finish!: (files: MayaAttachment[]) => void;
  const { queue, released } = fixture({
    pick: () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  });
  const pending = queue.pick();
  queue.background();
  finish([file()]);
  await pending;
  assert.equal(queue.getSnapshot().attachments.length, 1);
  queue.background();
  assert.deepEqual(queue.getSnapshot().attachments, []);
  assert.deepEqual(released, ["one"]);
});

test("removal while base64 is pending prevents the stale file from being sent", async () => {
  let finish!: (value: { size: number; base64: string }) => void;
  const { queue } = fixture({
    read: () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  });
  await queue.pick();
  const preparing = queue.prepare();
  queue.remove("one");
  finish({ size: 4, base64: "JVBERg==" });
  await assert.rejects(preparing, /conversation changed/);
  assert.equal(queue.getSnapshot().preparing, false);
});

test("failed reads retain attachments for an explicit retry; clearing releases copies", async () => {
  let fails = true;
  const { queue, released } = fixture({
    read: async () => {
      if (fails) throw new Error("Storage unavailable");
      return { size: 4, base64: "JVBERg==" };
    },
  });
  await queue.pick();
  await assert.rejects(queue.prepare(), /Storage unavailable/);
  assert.equal(queue.getSnapshot().attachments.length, 1);
  fails = false;
  assert.deepEqual(await queue.prepare(), [
    {
      type: "file",
      filename: "one.pdf",
      mediaType: "application/pdf",
      url: "data:application/pdf;base64,JVBERg==",
    },
  ]);
  queue.clear();
  assert.deepEqual(released, ["one"]);
});

test("a replacement session cannot admit a late picker result", async () => {
  let current = true;
  let finish!: (files: MayaAttachment[]) => void;
  const { queue, released } = fixture({
    isCurrent: () => current,
    pick: () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  });
  const pending = queue.pick();
  current = false;
  finish([file()]);
  await assert.rejects(pending, /conversation changed/);
  assert.deepEqual(released, ["one"]);
});
