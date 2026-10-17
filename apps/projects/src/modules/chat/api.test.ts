import type { JSONContent } from "@tiptap/core";
import {
  createMessage,
  editMessage,
  getMessage,
  getMessages,
  getResources,
} from "./api";
import type { ChatMessageInput } from "./types";

let mockResponse: unknown;
jest.mock("@/lib/http", () => {
  const request = (...args: unknown[]) => {
    const decode = args.at(-1);
    if (typeof decode !== "function")
      throw new Error("Expected a response decoder");
    return Promise.resolve(decode({ data: mockResponse }));
  };
  return {
    get: request,
    post: request,
    patch: request,
    put: request,
    remove: request,
  };
});
const context = { workspaceSlug: "workspace" };
const input: ChatMessageInput = {
  creationId: "creation-1",
  text: "https://fortyone.app",
  richContent: null,
  mentionedUserIds: [],
  agentMentions: [],
  attachmentIds: [],
};
const serverMessage = {
  id: "message-1",
  conversationId: "room-1",
  creationId: "creation-1",
  sequence: 1,
  author: { id: "author-1", fullName: "Sarah" },
  text: input.text,
  richContent: null as JSONContent | null,
  createdAt: "2026-10-17T09:00:00Z",
  version: 1,
};

describe("chat message link decoding", () => {
  it("preserves voice metadata across message and resource responses", async () => {
    const attachment = {
      id: "audio-1",
      fileName: "Voice note.m4a",
      mediaType: "audio/mp4",
      sizeBytes: 22567,
      kind: "voice-note",
      durationMs: 4375,
    };
    mockResponse = { ...serverMessage, attachments: [attachment] };
    expect(
      (await getMessage(context, "message-1")).attachments[0],
    ).toMatchObject(attachment);
    mockResponse = {
      items: [
        {
          ...attachment,
          id: "resource-1",
          attachmentId: attachment.id,
          attachmentKind: attachment.kind,
          kind: "file",
          conversationId: "room-1",
          messageId: "message-1",
          createdAt: serverMessage.createdAt,
          author: serverMessage.author,
        },
      ],
      nextCursor: null,
    };
    expect(
      (await getResources(context, "room-1", "files", "")).items[0],
    ).toMatchObject({ attachmentKind: "voice-note", durationMs: 4375 });
  });
  it("derives links missing from both history and newly created message responses", async () => {
    mockResponse = {
      items: [serverMessage],
      olderCursor: null,
      newerCursor: null,
      hasOlder: false,
      hasNewer: false,
    };
    const history = await getMessages(context, "room-1");
    expect(history.items[0].links.map((link) => link.url)).toEqual([
      "https://fortyone.app/",
    ]);
    mockResponse = serverMessage;
    const created = await createMessage(context, "room-1", input);
    expect(created.links).toEqual(history.items[0].links);
  });
  it("recomputes preview destinations from an edited rich document", async () => {
    const richContent: JSONContent = {
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            {
              type: "text",
              text: "https://label.example",
              marks: [
                {
                  type: "link",
                  attrs: { href: "https://destination.example/new#section" },
                },
              ],
            },
          ],
        },
      ],
    };
    mockResponse = {
      ...serverMessage,
      text: "https://label.example",
      richContent,
      version: 2,
    };
    const edited = await editMessage(context, "message-1", {
      ...input,
      richContent,
      version: 1,
    });
    expect(edited.links.map((link) => link.url)).toEqual([
      "https://destination.example/new#section",
    ]);
  });
  it("preserves server-provided link data, including an explicit empty list", async () => {
    const links = [
      { id: "persisted-link", url: "https://fortyone.app/", title: "FortyOne" },
    ];
    mockResponse = { ...serverMessage, links };
    expect((await getMessage(context, "message-1")).links).toEqual(links);
    mockResponse = { ...serverMessage, links: [] };
    expect((await getMessage(context, "message-1")).links).toEqual([]);
  });
  it("never exposes previews on deleted messages", async () => {
    mockResponse = {
      ...serverMessage,
      deletedAt: "2026-10-17T10:00:00Z",
      links: [{ id: "old", url: "https://fortyone.app/", title: "Old link" }],
    };
    expect((await getMessage(context, "message-1")).links).toEqual([]);
  });
});
