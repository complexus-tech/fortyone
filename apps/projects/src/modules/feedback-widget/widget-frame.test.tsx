/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ComponentPropsWithoutRef, ElementType, ReactNode } from "react";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import type { PublicPortal } from "@/modules/public-portal/types";
import {
  confirmWidgetFeedbackVerificationAction,
  createWidgetFeedbackAction,
  exchangeWidgetIdentityAction,
  markWidgetFeedbackUpdatesSeenAction,
  requestWidgetFeedbackVerificationAction,
  revokeWidgetIdentityAction,
  toggleWidgetFeedbackVoteAction,
} from "./actions";
import { FEEDBACK_WIDGET_CHANNEL } from "./protocol";
import { FeedbackWidgetFrame } from "./widget-frame";

jest.mock("./actions", () => ({
  confirmWidgetFeedbackVerificationAction: jest.fn(),
  createWidgetFeedbackAction: jest.fn(),
  createWidgetFeedbackCommentAction: jest.fn(),
  exchangeWidgetIdentityAction: jest.fn(),
  markWidgetFeedbackUpdatesSeenAction: jest.fn(),
  requestWidgetFeedbackVerificationAction: jest.fn(),
  revokeWidgetIdentityAction: jest.fn(),
  toggleWidgetFeedbackVoteAction: jest.fn(),
}));

jest.mock("icons", () => {
  const React = jest.requireActual("react");
  const Icon = (props: ComponentPropsWithoutRef<"svg">) =>
    React.createElement("svg", props);
  return new Proxy({}, { get: () => Icon });
});

jest.mock("lib", () => ({
  cn: (...values: unknown[]) => values.filter(Boolean).join(" "),
  getReadableTextColor: () => "#ffffff",
}));

jest.mock("ui", () => {
  const React = jest.requireActual("react");
  const Box = ({ children, ...props }: ComponentPropsWithoutRef<"div">) =>
    React.createElement("div", props, children);
  const Flex = Box;
  const Text = ({
    as = "span",
    children,
    color: _color,
    fontWeight: _fontWeight,
    ...props
  }: ComponentPropsWithoutRef<"span"> & {
    as?: ElementType;
    color?: string;
    fontWeight?: string;
  }) => React.createElement(as, props, children);
  const Button = ({
    asIcon: _asIcon,
    children,
    color: _color,
    leftIcon,
    rightIcon,
    size: _size,
    variant: _variant,
    ...props
  }: ComponentPropsWithoutRef<"button"> & {
    asIcon?: boolean;
    color?: string;
    leftIcon?: ReactNode;
    rightIcon?: ReactNode;
    size?: string;
    variant?: string;
  }) =>
    React.createElement(
      "button",
      { type: "button", ...props },
      leftIcon,
      children,
      rightIcon,
    );
  return {
    Avatar: ({ name, ...props }: { name: string }) =>
      React.createElement("div", { ...props, "aria-label": name }),
    Box,
    Button,
    Flex,
    Input: (props: ComponentPropsWithoutRef<"input">) =>
      React.createElement("input", props),
    Switch: ({
      checked,
      onCheckedChange,
      ...props
    }: ComponentPropsWithoutRef<"button"> & {
      checked: boolean;
      onCheckedChange: (checked: boolean) => void;
    }) =>
      React.createElement("button", {
        ...props,
        "aria-checked": checked,
        onClick: () => {
          onCheckedChange(!checked);
        },
        role: "switch",
        type: "button",
      }),
    Text,
  };
});

const exchangeIdentityMock = jest.mocked(exchangeWidgetIdentityAction);
const createFeedbackMock = jest.mocked(createWidgetFeedbackAction);
const requestVerificationMock = jest.mocked(
  requestWidgetFeedbackVerificationAction,
);
const confirmVerificationMock = jest.mocked(
  confirmWidgetFeedbackVerificationAction,
);
const markUpdatesSeenMock = jest.mocked(markWidgetFeedbackUpdatesSeenAction);
const revokeIdentityMock = jest.mocked(revokeWidgetIdentityAction);
const toggleVoteMock = jest.mocked(toggleWidgetFeedbackVoteAction);

const portal: PublicPortal = {
  boards: [
    {
      color: "blue",
      id: "board-1",
      name: "Product",
      slug: "product",
      teamId: "team-1",
    },
  ],
  guestIdentityPolicy: "allow_public_masking",
  hasPublishedUpdates: true,
  id: "portal-1",
  name: "Feedback",
  participationMode: "anonymous_allowed",
  requests: [],
  requestsHasMore: false,
  slug: "city-roads",
  updates: [],
  workspace: {
    avatarUrl: null,
    color: "#2563eb",
    name: "City Roads",
    slug: "city-roads",
  },
};

const request = {
  authorAvatar: null,
  authorId: null,
  authorName: "Amina",
  boardId: "board-1",
  commentCount: 0,
  comments: [],
  createdAtLabel: "Just now",
  description: "The crossing is unsafe.",
  id: "feedback-1",
  participantKind: "external" as const,
  slug: "repair-crossing",
  status: "pending" as const,
  storyLinks: [],
  title: "Repair the crossing",
  voteCount: 1,
};

const parentOrigin = "https://app.example.com";
let parentWindow: Window;
let parentPostMessage: ReturnType<typeof jest.fn>;

const renderWidget = (currentPortal = portal) =>
  render(
    <FeedbackWidgetFrame
      initialTab="feedback"
      instanceId="widget-1"
      mode="bubble"
      parentOrigin={parentOrigin}
      portal={currentPortal}
      roadmap={{ completed: [], in_progress: [], planned: [] }}
      theme="light"
      viewer={null}
    />,
  );

const sendHostIdentity = (assertion = "payload.signature") => {
  act(() => {
    window.dispatchEvent(
      new MessageEvent("message", {
        data: {
          channel: FEEDBACK_WIDGET_CHANNEL,
          event: "host-identify",
          instanceId: "widget-1",
          payload: { assertion, requestId: assertion },
          version: 1,
        },
        origin: parentOrigin,
        source: parentWindow,
      }),
    );
  });
};

const sendHostIdentityClear = () => {
  act(() => {
    window.dispatchEvent(
      new MessageEvent("message", {
        data: {
          channel: FEEDBACK_WIDGET_CHANNEL,
          event: "host-identity-clear",
          instanceId: "widget-1",
          payload: { requestId: "logout" },
          version: 1,
        },
        origin: parentOrigin,
        source: parentWindow,
      }),
    );
  });
};

const contributorSession = (token: string, unreadUpdateCount = 0) => ({
  data: {
    participant: {
      avatarUrl: null,
      canReceiveUpdates: true as const,
      displayName: "Amina",
      id: `contributor-${token}`,
      kind: "external" as const,
      masked: false,
      name: "Amina",
      sessionExpiresAt: "2026-08-13T00:00:00Z",
      unreadUpdateCount,
    },
    session: {
      expiresAt: "2026-08-13T00:00:00Z",
      token,
    },
  },
});

const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((nextResolve) => {
    resolve = nextResolve;
  });
  return { promise, resolve };
};

const writeDraft = () => {
  fireEvent.click(screen.getByRole("button", { name: "Add feedback" }));
  fireEvent.change(screen.getByLabelText("Feedback title"), {
    target: { value: "Repair the crossing" },
  });
};

describe("FeedbackWidgetFrame identity and guest participation", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    markUpdatesSeenMock.mockResolvedValue({
      data: {
        lastSeenAt: "2026-08-13T00:00:00Z",
        unreadUpdateCount: 0,
      },
    });
    revokeIdentityMock.mockResolvedValue({ data: null });
    toggleVoteMock.mockResolvedValue({
      data: { participantKind: "external", vote: 1, voteCount: 2 },
    });
    parentPostMessage = jest.fn();
    parentWindow = { postMessage: parentPostMessage } as unknown as Window;
    Object.defineProperty(window, "parent", {
      configurable: true,
      value: parentWindow,
    });
    Object.defineProperty(window, "matchMedia", {
      configurable: true,
      value: jest.fn(() => ({
        addEventListener: jest.fn(),
        matches: false,
        removeEventListener: jest.fn(),
      })),
    });
  });

  it("locks an identified submission to the iframe-only contributor session", async () => {
    exchangeIdentityMock.mockResolvedValue({
      data: {
        participant: {
          avatarUrl: null,
          canReceiveUpdates: true,
          displayName: "Amina",
          id: "contributor-1",
          kind: "external",
          masked: false,
          name: "Amina",
          sessionExpiresAt: "2026-08-13T00:00:00Z",
          unreadUpdateCount: 0,
        },
        session: {
          expiresAt: "2026-08-13T00:00:00Z",
          token: "iframe-session",
        },
      },
    });
    createFeedbackMock.mockResolvedValue({
      data: { participantKind: "external", request },
    });
    renderWidget();
    sendHostIdentity();

    await waitFor(() => {
      expect(exchangeIdentityMock).toHaveBeenCalled();
    });
    writeDraft();
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));

    await waitFor(() => {
      expect(createFeedbackMock).toHaveBeenCalledWith(
        expect.objectContaining({
          participationIntent: "external",
          sessionToken: "iframe-session",
          title: "Repair the crossing",
        }),
      );
    });
  });

  it("clears the prior session on logout without losing or misattributing the draft", async () => {
    exchangeIdentityMock.mockResolvedValue(contributorSession("session-a"));
    renderWidget();
    sendHostIdentity("identity-a");

    expect(
      await screen.findByRole("button", { name: "View feedback updates" }),
    ).toBeInTheDocument();
    writeDraft();
    sendHostIdentityClear();

    expect(screen.getByLabelText("Feedback title")).toHaveValue(
      "Repair the crossing",
    );
    expect(revokeIdentityMock).toHaveBeenCalledWith({
      portalSlug: "city-roads",
      sessionToken: "session-a",
    });
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));
    expect(
      screen.getByRole("button", { name: "Submit anonymously" }),
    ).toBeInTheDocument();
    expect(createFeedbackMock).not.toHaveBeenCalled();
    expect(parentPostMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        event: "identity-cleared",
        payload: { requestId: "logout" },
      }),
      parentOrigin,
    );
  });

  it("locks all writes during an A-to-B switch and resumes with only B", async () => {
    const identityB = deferred<ReturnType<typeof contributorSession>>();
    exchangeIdentityMock.mockImplementation((input) =>
      input.assertion === "identity-b"
        ? identityB.promise
        : Promise.resolve(contributorSession("session-a")),
    );
    createFeedbackMock.mockResolvedValue({
      data: { participantKind: "external", request },
    });
    renderWidget({ ...portal, requests: [request] });
    sendHostIdentity("identity-a");
    expect(
      await screen.findByRole("button", { name: "View feedback updates" }),
    ).toBeInTheDocument();
    writeDraft();

    sendHostIdentity("identity-b");

    expect(
      screen.getByRole("button", { name: "Submit feedback" }),
    ).toBeDisabled();
    expect(
      screen.getByRole("button", { name: "Upvote feedback" }),
    ).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));
    fireEvent.click(screen.getByRole("button", { name: "Upvote feedback" }));
    expect(createFeedbackMock).not.toHaveBeenCalled();
    expect(toggleVoteMock).not.toHaveBeenCalled();
    expect(revokeIdentityMock).toHaveBeenCalledWith({
      portalSlug: "city-roads",
      sessionToken: "session-a",
    });

    await act(async () => {
      identityB.resolve(contributorSession("session-b"));
      await identityB.promise;
    });
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Submit feedback" }),
      ).not.toBeDisabled();
    });
    expect(screen.getByLabelText("Feedback title")).toHaveValue(
      "Repair the crossing",
    );
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));
    await waitFor(() => {
      expect(createFeedbackMock).toHaveBeenCalledWith(
        expect.objectContaining({
          participationIntent: "external",
          sessionToken: "session-b",
          title: "Repair the crossing",
        }),
      );
    });
    expect(
      screen.queryByRole("button", { name: "Submit anonymously" }),
    ).not.toBeInTheDocument();
  });

  it("ignores and revokes a stale identity exchange response", async () => {
    const identityA = deferred<ReturnType<typeof contributorSession>>();
    exchangeIdentityMock.mockImplementation((input) =>
      input.assertion === "identity-a"
        ? identityA.promise
        : Promise.resolve(contributorSession("session-b")),
    );
    renderWidget({ ...portal, requests: [request] });
    sendHostIdentity("identity-a");
    sendHostIdentity("identity-b");
    await waitFor(() => {
      expect(parentPostMessage).toHaveBeenCalledWith(
        expect.objectContaining({
          event: "identity-ready",
          payload: expect.objectContaining({ requestId: "identity-b" }),
        }),
        parentOrigin,
      );
    });

    await act(async () => {
      identityA.resolve(contributorSession("stale-session-a"));
      await identityA.promise;
    });
    expect(revokeIdentityMock).toHaveBeenCalledWith({
      portalSlug: "city-roads",
      sessionToken: "stale-session-a",
    });
    fireEvent.click(screen.getByRole("button", { name: "Upvote feedback" }));
    await waitFor(() => {
      expect(toggleVoteMock).toHaveBeenCalledWith(
        expect.objectContaining({ sessionToken: "session-b" }),
      );
    });
  });

  it("initializes the unread badge and clears it when Updates opens", async () => {
    exchangeIdentityMock.mockResolvedValue(contributorSession("session-a", 3));
    renderWidget();
    sendHostIdentity("identity-a");

    const updatesButton = await screen.findByRole("button", {
      name: "Updates, 3 unread updates",
    });
    fireEvent.click(updatesButton);

    await waitFor(() => {
      expect(markUpdatesSeenMock).toHaveBeenCalledWith({
        portalSlug: "city-roads",
        sessionToken: "session-a",
      });
      expect(
        screen.getByRole("button", { name: "Updates" }),
      ).toBeInTheDocument();
    });
  });

  it("keeps unread updates at zero for anonymous visitors", () => {
    renderWidget();

    expect(screen.getByRole("button", { name: "Updates" })).toBeInTheDocument();
    expect(markUpdatesSeenMock).not.toHaveBeenCalled();
  });

  it("requires an explicit anonymous choice for a contactless submission", async () => {
    createFeedbackMock.mockResolvedValue({
      data: {
        participantKind: "anonymous",
        request: {
          ...request,
          authorName: "Anonymous",
          participantKind: "anonymous",
        },
      },
    });
    renderWidget();
    writeDraft();
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));

    expect(
      screen.getByRole("button", { name: "Submit anonymously" }),
    ).toBeInTheDocument();
    expect(createFeedbackMock).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Submit anonymously" }));

    await waitFor(() => {
      expect(createFeedbackMock).toHaveBeenCalledWith(
        expect.objectContaining({ participationIntent: "anonymous" }),
      );
    });
  });

  it("preserves the draft through email-code verification", async () => {
    requestVerificationMock.mockResolvedValue({
      data: { accepted: true, expiresAt: "2026-08-12T01:00:00Z" },
    });
    confirmVerificationMock.mockResolvedValue({
      data: {
        participant: {
          avatarUrl: null,
          canReceiveUpdates: true,
          displayName: "Amina",
          email: "amina@example.com",
          id: "contributor-1",
          kind: "verified_guest",
          masked: false,
          name: "Amina",
          sessionExpiresAt: "2026-09-11T00:00:00Z",
          unreadUpdateCount: 0,
        },
        session: {
          expiresAt: "2026-09-11T00:00:00Z",
          token: "guest-session",
        },
      },
    });
    createFeedbackMock.mockResolvedValue({
      data: {
        participantKind: "verified_guest",
        request: { ...request, participantKind: "verified_guest" },
      },
    });
    renderWidget({ ...portal, participationMode: "verified_guest" });
    writeDraft();
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));
    fireEvent.click(
      screen.getByRole("button", { name: "Continue with email" }),
    );
    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "amina@example.com" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Send verification code" }),
    );

    expect(await screen.findByText("Check your email")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Verification code"), {
      target: { value: "123456" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Verify and continue" }),
    );
    await waitFor(() => {
      expect(screen.getByLabelText("Feedback title")).toHaveValue(
        "Repair the crossing",
      );
    });
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));

    await waitFor(() => {
      expect(createFeedbackMock).toHaveBeenCalledWith(
        expect.objectContaining({
          participationIntent: "verified_guest",
          sessionToken: "guest-session",
          title: "Repair the crossing",
        }),
      );
    });
  });
});
