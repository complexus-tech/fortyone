import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { useTeamMembersInfinite } from "@/lib/hooks/team-members";
import type { Member } from "@/types/member";
import { ChatIdentityProvider } from "./chat-context";
import { ConversationMembersPopover } from "./conversation-members-popover";
import type { ChatConversation, ChatPerson } from "./types";

jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembersInfinite: jest.fn(),
}));
jest.mock("@/lib/hooks/members", () => ({ MEMBER_MENU_PAGE_SIZE: 15 }));

const person = (
  id: string,
  fullName: string,
  isSystem = false,
): ChatPerson => ({
  id,
  fullName,
  isSystem,
  avatarUrl: null,
});
const maya = person("maya", "Maya", true);
const sarah = person("sarah", "Sarah Chen");
const daniel = person("daniel", "Daniel Brooks");
const asMember = (member: ChatPerson, active = true): Member => ({
  ...member,
  username: member.fullName,
  email: "",
  role: "member",
  isActive: active,
  isInternal: false,
  createdAt: "",
  updatedAt: "",
});
const conversation: ChatConversation = {
  id: "product-conversation",
  teamId: "product-team",
  kind: "team",
  name: "Product",
  description: "",
  color: null,
  members: [maya, sarah],
  memberCount: 3,
  unreadCount: 0,
  unreadRootCount: 0,
  unreadReplyCount: 0,
  lastReadSequence: 0,
  lastMessageAt: null,
  lastMessageText: "",
};
const useMembersMock = jest.mocked(useTeamMembersInfinite);
const queryResult = {
  data: {
    pages: [{ members: [asMember(sarah), asMember(daniel), asMember(maya)] }],
  },
  isPending: false,
  isError: false,
  isFetching: false,
  isFetchingNextPage: false,
  isFetchNextPageError: false,
  hasNextPage: false,
  fetchNextPage: jest.fn(),
  refetch: jest.fn(),
};

const setQueryResult = (overrides: Record<string, unknown> = {}) => {
  useMembersMock.mockReturnValue({
    ...queryResult,
    ...overrides,
  } as unknown as ReturnType<typeof useTeamMembersInfinite>);
};

const openMembers = () => {
  render(
    <ChatIdentityProvider
      value={{
        userId: sarah.id,
        conversationId: conversation.id,
        members: conversation.members,
        mayaEnabled: true,
      }}
    >
      <ConversationMembersPopover conversation={conversation} />
    </ChatIdentityProvider>,
  );
  expect(useMembersMock).not.toHaveBeenCalled();
  const trigger = screen.getByRole("button", {
    name: "View Product team members",
  });
  fireEvent.click(trigger);
  return trigger;
};

beforeEach(() => {
  jest.clearAllMocks();
  setQueryResult();
});

describe("ConversationMembersPopover", () => {
  it("opens the full team roster on demand using shared popover focus behavior", async () => {
    const trigger = openMembers();
    const dialog = screen.getByRole("dialog", { name: "Team members" });
    const members = within(dialog).getByRole("list", { name: "Members" });
    expect(within(members).getByText("Daniel Brooks")).toBeInTheDocument();
    expect(within(members).getAllByText("Maya")).toHaveLength(1);
    expect(within(members).getByText("You")).toBeInTheDocument();
    expect(useMembersMock).toHaveBeenLastCalledWith("product-team", "", 15);
    expect(
      screen.getByRole("textbox", { name: "Search team members" }),
    ).toHaveFocus();

    fireEvent.keyDown(document.activeElement!, { key: "Escape" });
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
      expect(trigger).toHaveFocus();
    });
  });

  it("shows known members immediately while the full roster loads", () => {
    setQueryResult({ data: undefined, isPending: true, isFetching: true });
    openMembers();
    const members = screen.getByRole("list", { name: "Members" });
    expect(within(members).getByText("Sarah Chen")).toBeInTheDocument();
    expect(within(members).getByText("Maya")).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("Loading members");
  });

  it("filters Maya as well as people and searches the paginated team roster", async () => {
    openMembers();
    fireEvent.change(
      screen.getByRole("textbox", { name: "Search team members" }),
      {
        target: { value: "maya" },
      },
    );
    const members = screen.getByRole("list", { name: "Members" });
    expect(within(members).getAllByRole("listitem")).toHaveLength(1);
    expect(within(members).getByText("Maya")).toBeInTheDocument();
    await waitFor(() => {
      expect(useMembersMock).toHaveBeenLastCalledWith(
        "product-team",
        "maya",
        15,
      );
    });
  });

  it("deduplicates members across pages and excludes inactive members", () => {
    setQueryResult({
      data: {
        pages: [
          { members: [asMember(sarah), asMember(daniel)] },
          {
            members: [
              asMember(daniel),
              asMember(person("inactive", "Inactive"), false),
            ],
          },
        ],
      },
    });
    openMembers();
    const members = screen.getByRole("list", { name: "Members" });
    expect(within(members).getAllByText("Daniel Brooks")).toHaveLength(1);
    expect(within(members).queryByText("Inactive")).not.toBeInTheDocument();
  });

  it("supports loading more without a pointer and exposes retry after failures", () => {
    setQueryResult({ hasNextPage: true });
    const { unmount } = render(
      <ConversationMembersPopover conversation={conversation} />,
    );
    fireEvent.click(
      screen.getByRole("button", { name: "View Product team members" }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Load more members" }));
    expect(queryResult.fetchNextPage).toHaveBeenCalledTimes(1);
    unmount();

    setQueryResult({ data: undefined, isError: true });
    render(<ConversationMembersPopover conversation={conversation} />);
    fireEvent.click(
      screen.getByRole("button", { name: "View Product team members" }),
    );
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Could not load members.",
    );
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(queryResult.refetch).toHaveBeenCalledTimes(1);
  });
});
