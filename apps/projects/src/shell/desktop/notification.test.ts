import { getDesktopBridge } from "./bridge";
import { bindDesktopSession } from "./session-binding";
import {
  deliverDesktopNotification,
  parseDesktopNotification,
} from "./notification";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./session-binding", () => ({ bindDesktopSession: jest.fn() }));

const scope = {
  accountId: "11111111-1111-4111-8111-111111111111",
  workspaceId: "22222222-2222-4222-8222-222222222222",
  workspaceSlug: "engineering",
};
const event = {
  id: "33333333-3333-4333-8333-333333333333",
  recipient_id: scope.accountId,
  workspace_id: scope.workspaceId,
  entity_type: "story",
  entity_id: "44444444-4444-4444-8444-444444444444",
  title: "New comment",
  type: "story_comment",
  message: {
    template: "Ada: {content}",
    variables: { content: { value: "<p>Please review</p>", type: "text" } },
  },
  created_at: "2026-10-15T12:00:00.123456Z",
  read_at: null,
};
const deliverNotification = jest.fn();

beforeEach(() => {
  jest.clearAllMocks();
  jest
    .mocked(getDesktopBridge)
    .mockReturnValue({ deliverNotification } as unknown as NonNullable<
      ReturnType<typeof getDesktopBridge>
    >);
  jest.mocked(bindDesktopSession).mockResolvedValue(true);
  deliverNotification.mockResolvedValue(undefined);
});

it("normalizes actual SSE payloads without losing coalesced notification precision", () => {
  expect(parseDesktopNotification(JSON.stringify(event), scope)).toEqual({
    ...scope,
    notificationId: event.id,
    createdAt: event.created_at,
    entityType: "story",
    entityId: event.entity_id,
    title: "New comment",
    body: "Ada: Please review",
  });
});

it.each([
  { recipient_id: "55555555-5555-4555-8555-555555555555" },
  { workspace_id: "55555555-5555-4555-8555-555555555555" },
  { recipientId: "55555555-5555-4555-8555-555555555555" },
  { read_at: "2026-10-15T12:01:00Z" },
  { entity_type: "https://untrusted.example" },
  { created_at: "yesterday" },
  { message: { template: "Hi", variables: { name: { value: {} } } } },
])(
  "does not forward an out-of-scope or malformed notification %j",
  async (patch) => {
    await deliverDesktopNotification(
      JSON.stringify({ ...event, ...patch }),
      scope,
      () => true,
    );
    expect(bindDesktopSession).not.toHaveBeenCalled();
    expect(deliverNotification).not.toHaveBeenCalled();
  },
);

it("binds the account before native delivery and does not suppress a focused window", async () => {
  await deliverDesktopNotification(JSON.stringify(event), scope, () => true);
  expect(bindDesktopSession).toHaveBeenCalledWith(scope);
  expect(deliverNotification).toHaveBeenCalledWith(
    expect.objectContaining({ notificationId: event.id }),
  );
});

it("drops a notification when the workspace changes while binding", async () => {
  let finish!: (bound: boolean) => void;
  let current = true;
  jest.mocked(bindDesktopSession).mockReturnValue(
    new Promise((resolve) => {
      finish = resolve;
    }),
  );
  const pending = deliverDesktopNotification(
    JSON.stringify(event),
    scope,
    () => current,
  );
  current = false;
  finish(true);
  await pending;
  expect(deliverNotification).not.toHaveBeenCalled();
});

it("does nothing in browsers and older desktop builds", async () => {
  jest.mocked(getDesktopBridge).mockReturnValue(null);
  await deliverDesktopNotification(JSON.stringify(event), scope, () => true);
  expect(bindDesktopSession).not.toHaveBeenCalled();
});

it("retains a validated comment target in an incoming notification", () => {
  const commentId = "55555555-5555-4555-8555-555555555555";
  const notification = parseDesktopNotification(
    JSON.stringify({
      ...event,
      message: {
        ...event.message,
        variables: {
          ...event.message.variables,
          commentId: { value: commentId, type: "comment" },
        },
      },
    }),
    scope,
  );
  expect(notification?.commentId).toBe(commentId);
  expect(notification?.body).toBe("Ada: Please review");
});

it.each(["not-a-uuid", "00000000-0000-0000-0000-000000000000"])(
  "ignores invalid comment targets without dropping the notification (%s)",
  (commentId) => {
    const notification = parseDesktopNotification(
      JSON.stringify({
        ...event,
        message: {
          ...event.message,
          variables: { commentId: { value: commentId } },
        },
      }),
      scope,
    );
    expect(notification).not.toBeNull();
    expect(notification).not.toHaveProperty("commentId");
  },
);

it.each([false, true])(
  "only forwards comment metadata to capable native builds (%s)",
  async (supported) => {
    const commentId = "55555555-5555-4555-8555-555555555555";
    jest.mocked(getDesktopBridge).mockReturnValue({
      deliverNotification,
      ...(supported ? { commentNotificationLinks: true } : {}),
    } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);
    await deliverDesktopNotification(
      JSON.stringify({
        ...event,
        message: {
          ...event.message,
          variables: { commentId: { value: commentId } },
        },
      }),
      scope,
      () => true,
    );
    expect(deliverNotification).toHaveBeenCalledTimes(1);
    const notification = deliverNotification.mock.calls[0][0];
    expect(notification.commentId).toBe(supported ? commentId : undefined);
  },
);
