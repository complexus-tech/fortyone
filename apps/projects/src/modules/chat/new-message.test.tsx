/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */

import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { Member } from "@/types/member";
import { createDirectConversation } from "./api";
import { chatKeys } from "./constants";
import { NewMessage } from "./new-message";
import type { ChatConversation } from "./types";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "second" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "me" } } }),
}));
jest.mock("@/lib/hooks/members", () => ({
  useMembersInfinite: (...args: unknown[]) => mockPeopleQuery(...args),
}));
jest.mock("./queries", () => ({
  useChatConversations: (...args: unknown[]) => mockTeamsQuery(...args),
}));
jest.mock("./api", () => ({ createDirectConversation: jest.fn() }));
jest.mock("./composer", () => ({
  Composer: () => <textarea aria-label="Start a new message" />,
}));
jest.mock("./avatars", () => ({ ConversationMark: () => <span>#</span> }));

const mockPeopleQuery = jest.fn();
const mockTeamsQuery = jest.fn();
const member: Member = {
  id: "sarah",
  fullName: "Sarah Chen",
  username: "sarah",
  email: "sarah@example.com",
  avatarUrl: null,
  isActive: true,
  isSystem: false,
  isInternal: false,
  role: "member",
  createdAt: "2026-10-17T00:00:00Z",
  updatedAt: "2026-10-17T00:00:00Z",
};
const team: ChatConversation = {
  id: "product-chat",
  kind: "team",
  teamId: "product",
  name: "Product",
  description: "",
  color: null,
  members: [],
  memberCount: 3,
  unreadCount: 0,
  lastReadSequence: 0,
  lastMessageAt: null,
  lastMessageText: "",
};
const direct: ChatConversation = {
  ...team,
  id: "sarah-chat",
  kind: "dm",
  teamId: null,
  name: "Sarah Chen",
};
const queryState = () => ({
  isPending: false,
  isError: false,
  hasNextPage: false,
  isFetchingNextPage: false,
  fetchNextPage: jest.fn().mockResolvedValue(undefined),
  refetch: jest.fn().mockResolvedValue(undefined),
});
let client: QueryClient;
const onSelect = jest.fn();
const onBack = jest.fn();
const create = jest.mocked(createDirectConversation);
const mount = () =>
  render(
    <QueryClientProvider client={client}>
      <NewMessage
        draft={{ text: "", content: null }}
        onBack={onBack}
        onDraftChange={jest.fn()}
        onSelect={onSelect}
      />
    </QueryClientProvider>,
  );

beforeEach(() => {
  jest.clearAllMocks();
  client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  Element.prototype.scrollIntoView = jest.fn();
  mockPeopleQuery.mockReturnValue({
    ...queryState(),
    data: {
      pages: [
        {
          members: [
            member,
            { ...member, id: "me", fullName: "Me" },
            { ...member, id: "maya", isSystem: true },
            { ...member, id: "inactive", isActive: false },
          ],
        },
      ],
    },
  });
  mockTeamsQuery.mockReturnValue({
    ...queryState(),
    data: { pages: [{ items: [team, direct] }] },
  });
  create.mockResolvedValue(direct);
});

afterEach(() => {
  cleanup();
  client.clear();
});

describe("new message recipients", () => {
  it("searches people and teams, keeps keyboard focus in To, and opens the selected team", () => {
    mount();
    const input = screen.getByRole("combobox", {
      name: "Find a person or team",
    });
    expect(input).toHaveFocus();
    expect(screen.getAllByRole("option")).toHaveLength(2);
    expect(mockPeopleQuery).toHaveBeenCalledWith("", 15, true);
    expect(mockTeamsQuery).toHaveBeenCalledWith("");
    fireEvent.keyDown(input, { key: "ArrowDown" });
    expect(screen.getByRole("option", { name: /Product/ })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(input).toHaveFocus();
    fireEvent.keyDown(input, { key: "Enter" });
    expect(onSelect).toHaveBeenCalledWith(team.id);
    expect(create).not.toHaveBeenCalled();
  });

  it("closes and reopens suggestions with Escape and ArrowDown without abandoning the draft", () => {
    mount();
    const input = screen.getByRole("combobox");
    fireEvent.keyDown(input, { key: "Escape" });
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    expect(input).toHaveFocus();
    expect(onBack).not.toHaveBeenCalled();
    fireEvent.keyDown(input, { key: "ArrowDown" });
    expect(screen.getByRole("listbox")).toBeInTheDocument();
    expect(screen.getByRole("option", { name: /Sarah/ })).toHaveAttribute(
      "aria-selected",
      "true",
    );
  });

  it("does not select stale results while a new search is debouncing", async () => {
    mount();
    const input = screen.getByRole("combobox");
    fireEvent.change(input, { target: { value: "Product" } });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(onSelect).not.toHaveBeenCalled();
    expect(create).not.toHaveBeenCalled();
    await waitFor(() => {
      expect(mockTeamsQuery).toHaveBeenLastCalledWith("Product");
    });
    expect(mockPeopleQuery).toHaveBeenLastCalledWith("Product", 15, true);
  });

  it("prevents duplicate direct-message requests and caches the opened conversation", async () => {
    let resolve: (conversation: ChatConversation) => void = () => undefined;
    create.mockReturnValue(
      new Promise((done) => {
        resolve = done;
      }),
    );
    mount();
    const option = screen.getByRole("option", { name: /Sarah/ });
    fireEvent.click(option);
    fireEvent.click(option);
    await waitFor(() => {
      expect(create).toHaveBeenCalledTimes(1);
    });
    await act(async () => {
      resolve(direct);
    });
    await waitFor(() => {
      expect(onSelect).toHaveBeenCalledWith(direct.id);
    });
    expect(
      client.getQueryData(chatKeys.conversation("second", direct.id)),
    ).toEqual(direct);
  });

  it.each(["back", "unmount"])(
    "does not reopen a conversation after %s while creation is pending",
    async (action) => {
      let resolve: (conversation: ChatConversation) => void = () => undefined;
      create.mockReturnValue(
        new Promise((done) => {
          resolve = done;
        }),
      );
      const view = mount();
      fireEvent.click(screen.getByRole("option", { name: /Sarah/ }));
      await waitFor(() => {
        expect(create).toHaveBeenCalledTimes(1);
      });
      if (action === "back")
        fireEvent.click(
          screen.getByRole("button", { name: "Back to conversations" }),
        );
      else view.unmount();
      await act(async () => {
        resolve(direct);
      });
      expect(onSelect).not.toHaveBeenCalled();
    },
  );

  it("shows a recoverable direct-message failure and retries the same person", async () => {
    create.mockRejectedValueOnce(new Error("Connection interrupted"));
    mount();
    fireEvent.click(screen.getByRole("option", { name: /Sarah/ }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Connection interrupted",
    );
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    await waitFor(() => {
      expect(onSelect).toHaveBeenCalledWith(direct.id);
    });
    expect(create).toHaveBeenNthCalledWith(
      2,
      { workspaceSlug: "second" },
      member.id,
    );
  });

  it("keeps team results available when people fail and offers independent retries and pagination", () => {
    const peopleState = { ...queryState(), isError: true };
    const teamsState = {
      ...queryState(),
      hasNextPage: true,
      data: { pages: [{ items: [team] }] },
    };
    mockPeopleQuery.mockReturnValue(peopleState);
    mockTeamsQuery.mockReturnValue(teamsState);
    mount();
    expect(screen.getByRole("option", { name: /Product/ })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Retry people" }));
    expect(peopleState.refetch).toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Load more teams" }));
    expect(teamsState.fetchNextPage).toHaveBeenCalled();
  });
});
