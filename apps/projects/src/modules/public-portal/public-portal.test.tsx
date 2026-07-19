/* global beforeAll, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactTypes from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { publicPortalFixture } from "./fixtures";
import {
  createFeedbackAction,
  createFeedbackCommentAction,
  toggleFeedbackVoteAction,
} from "./actions";
import {
  PublicPortalRequestDetailPage,
  PublicPortalRequestsPage,
  PublicPortalRoadmapPage,
} from ".";

const clipboardWriteTextMock = jest.fn(async (_text: string) => undefined);
const shareMock = jest.fn(async (_data?: ShareData) => undefined);
const portalViewer = {
  accountHref: "/city-roads/settings/account",
  appHref: "/city-roads/my-work",
  avatarUrl: null,
  email: "ada@example.com",
  name: "Ada Ndlovu",
  notificationsHref: "/city-roads/notifications",
};

const designSystemOnlyProps = new Set([
  "active",
  "align",
  "asIcon",
  "color",
  "direction",
  "fontSize",
  "fontWeight",
  "fullWidth",
  "gap",
  "justify",
  "rounded",
  "size",
  "variant",
  "wrap",
]);

const getDomProps = <T extends Record<string, unknown>>(props: T) =>
  Object.fromEntries(
    Object.entries(props).filter(([key]) => !designSystemOnlyProps.has(key)),
  );

jest.mock("icons", () => {
  const Icon = () => <svg aria-hidden="true" />;
  return {
    ArrowRightIcon: Icon,
    ArrowRight2Icon: Icon,
    ArrowUpDownIcon: Icon,
    ArrowUpIcon: Icon,
    BellIcon: Icon,
    CheckIcon: Icon,
    CommentIcon: Icon,
    CopyIcon: Icon,
    DashboardIcon: Icon,
    GanttIcon: Icon,
    KanbanIcon: Icon,
    ListIcon: Icon,
    LogoutIcon: Icon,
    MoonIcon: Icon,
    PlusIcon: Icon,
    RequestsIcon: Icon,
    RoadmapIcon: Icon,
    SearchIcon: Icon,
    ShareIcon: Icon,
    SettingsIcon: Icon,
    StoryIcon: Icon,
    SunIcon: Icon,
    SystemIcon: Icon,
    ThumbsDownIcon: Icon,
    ThumbsUpIcon: Icon,
    UpdatesIcon: Icon,
  };
});

jest.mock("next-themes", () => ({
  useTheme: () => ({
    setTheme: jest.fn(),
    theme: "system",
  }),
}));

jest.mock("@/components/shared/sidebar/actions", () => ({
  logOut: jest.fn(),
}));

jest.mock("@/components/shared/sidebar/utils", () => ({
  clearAllStorage: jest.fn(),
}));

jest.mock("./actions", () => ({
  createFeedbackAction: jest.fn(),
  createFeedbackCommentAction: jest.fn(),
  toggleFeedbackVoteAction: jest.fn(),
}));

const createFeedbackActionMock = jest.mocked(createFeedbackAction);
const createFeedbackCommentActionMock = jest.mocked(
  createFeedbackCommentAction,
);
const toggleFeedbackVoteActionMock = jest.mocked(toggleFeedbackVoteAction);

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
}));

jest.mock("ui", () => {
  const React = jest.requireActual("react");

  const Box = ({
    children,
    ...props
  }: ReactTypes.HTMLAttributes<HTMLDivElement> & Record<string, unknown>) => (
    <div {...getDomProps(props)}>{children}</div>
  );
  const Flex = ({
    children,
    ...props
  }: ReactTypes.HTMLAttributes<HTMLDivElement> & Record<string, unknown>) => (
    <div {...getDomProps(props)}>{children}</div>
  );
  const Text = ({
    as: Tag = "p",
    children,
    ...props
  }: ReactTypes.HTMLAttributes<HTMLElement> &
    Record<string, unknown> & {
      as?: ReactTypes.ElementType;
    }) => React.createElement(Tag, getDomProps(props), children);
  const Button = ({
    children,
    href,
    leftIcon,
    rightIcon,
    ...props
  }: ReactTypes.ButtonHTMLAttributes<HTMLButtonElement> &
    Record<string, unknown> & {
      href?: string;
      leftIcon?: ReactTypes.ReactNode;
      rightIcon?: ReactTypes.ReactNode;
    }) =>
    href ? (
      <a {...getDomProps(props)} href={href}>
        {leftIcon}
        {children}
        {rightIcon}
      </a>
    ) : (
      <button {...getDomProps(props)} type="button">
        {leftIcon}
        {children}
        {rightIcon}
      </button>
    );
  const Avatar = ({ name }: { name?: string }) => (
    <div>{name ? name.slice(0, 2).toUpperCase() : "U"}</div>
  );
  const Dialog = ({
    children,
    open,
  }: {
    children: ReactTypes.ReactNode;
    open?: boolean;
  }) => <div>{open ? children : children}</div>;
  function DialogContent({
    children,
    hideClose,
  }: {
    children: ReactTypes.ReactNode;
    hideClose?: boolean;
  }) {
    return (
      <div>
        {children}
        {hideClose ? null : <button type="button">Close</button>}
      </div>
    );
  }
  function DialogHeader({ children }: { children: ReactTypes.ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogTitle({ children }: { children: ReactTypes.ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogBody({ children }: { children: ReactTypes.ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogFooter({ children }: { children: ReactTypes.ReactNode }) {
    return <div>{children}</div>;
  }
  const DialogClose = () => <button type="button">Close</button>;
  Dialog.Content = DialogContent;
  Dialog.Header = DialogHeader;
  Dialog.Title = DialogTitle;
  Dialog.Body = DialogBody;
  Dialog.Footer = DialogFooter;
  Dialog.Close = DialogClose;
  const Input = ({
    leftIcon,
    ...props
  }: ReactTypes.InputHTMLAttributes<HTMLInputElement> &
    Record<string, unknown> & {
      leftIcon?: ReactTypes.ReactNode;
    }) => {
    void leftIcon;
    return <input {...getDomProps(props)} />;
  };
  const TextArea = (
    props: ReactTypes.TextareaHTMLAttributes<HTMLTextAreaElement>,
  ) => <textarea {...props} />;
  const TextEditor = ({
    editor,
    ...props
  }: ReactTypes.TextareaHTMLAttributes<HTMLTextAreaElement> & {
    editor?: { commands?: { setContent: (value: string) => void } };
  }) => (
    <textarea
      data-testid="text-editor"
      {...props}
      onChange={(event) => {
        editor?.commands?.setContent(event.target.value);
      }}
    />
  );
  const Menu = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuButton = ({ children }: { children: ReactTypes.ReactNode }) => (
    <>{children}</>
  );
  const MenuItems = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuGroup = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuSeparator = () => <hr />;
  const MenuItem = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuSubMenu = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuSubTrigger = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuSubItems = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  Menu.Button = MenuButton;
  Menu.Items = MenuItems;
  Menu.Group = MenuGroup;
  Menu.Separator = MenuSeparator;
  Menu.Item = MenuItem;
  Menu.SubMenu = MenuSubMenu;
  Menu.SubTrigger = MenuSubTrigger;
  Menu.SubItems = MenuSubItems;

  return {
    Avatar,
    Box,
    Button,
    Dialog,
    Flex,
    Input,
    Menu,
    Text,
    TextArea,
    TextEditor,
  };
});

describe("Public portal UI", () => {
  beforeAll(() => {
    class IntersectionObserverMock {
      disconnect = jest.fn();
      observe = jest.fn();
      unobserve = jest.fn();
    }

    Object.defineProperty(global, "IntersectionObserver", {
      value: IntersectionObserverMock,
      writable: true,
    });
  });

  beforeEach(() => {
    jest.clearAllMocks();
    createFeedbackActionMock.mockResolvedValue({ data: null });
    createFeedbackCommentActionMock.mockResolvedValue({
      data: {
        authorAvatar: null,
        authorName: "Ada Ndlovu",
        body: "Thanks for raising this.",
        createdAt: new Date().toISOString(),
        id: "comment-new",
      },
    });
    toggleFeedbackVoteActionMock.mockImplementation(async ({ vote }) => ({
      data: { vote, voteCount: vote === 1 ? 13 : 11 },
    }));
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText: clipboardWriteTextMock },
    });
    Object.defineProperty(navigator, "share", {
      configurable: true,
      value: shareMock,
    });
    global.fetch = jest.fn(
      async (input: Parameters<typeof fetch>[0]): Promise<Response> => {
        let url: string;
        if (typeof input === "string") {
          url = input;
        } else if (input instanceof URL) {
          url = input.toString();
        } else {
          url = input.url;
        }
        const requestUrl = new URL(url, "https://fortyone.test");
        const status = requestUrl.searchParams.get("status");
        const boardId = requestUrl.searchParams.get("boardId");
        const requests = publicPortalFixture.requests.filter(
          (request) =>
            (!status || request.status === status) &&
            (!boardId || request.boardId === boardId),
        );

        return {
          json: async () => ({
            data: {
              ...publicPortalFixture,
              requests,
              requestsHasMore: false,
            },
          }),
          ok: true,
        } as Response;
      },
    );
  });

  it("renders the public feedback page with feedback terminology", () => {
    render(
      <PublicPortalRequestsPage
        portal={publicPortalFixture}
        viewer={portalViewer}
      />,
    );

    expect(
      screen.getAllByRole("link", { name: /^Feedback$/i }).length,
    ).toBeGreaterThan(0);
    expect(
      screen.getByRole("button", { name: /new feedback/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("All boards")).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "Login/signup" }),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("All Requests")).not.toBeInTheDocument();
    expect(screen.getByTestId("text-editor")).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Close" })).toHaveLength(1);
  });

  it("sends logged-out visitors to login before submitting feedback", () => {
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    expect(
      screen.getByRole("link", { name: "Login to submit feedback" }),
    ).toHaveAttribute("href", "/");
    expect(
      screen.queryByRole("button", { name: "New Feedback" }),
    ).not.toBeInTheDocument();
  });

  it("filters public feedback by the selected board", async () => {
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    expect(screen.getByText("All boards")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Drainage" }));

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("boardId=drainage"),
      );
    });
    expect(
      await screen.findByText("Blocked storm drain on 4th Street"),
    ).toBeVisible();
    await waitFor(() => {
      expect(
        screen.queryByText("Add pedestrian crossing near East Avenue school"),
      ).not.toBeInTheDocument();
    });
  });

  it("automatically selects the only board when submitting feedback", async () => {
    const portal = {
      ...publicPortalFixture,
      boards: [publicPortalFixture.boards[0]],
    };
    render(<PublicPortalRequestsPage portal={portal} viewer={portalViewer} />);

    fireEvent.change(screen.getByLabelText("Feedback title"), {
      target: { value: "Add a safer crossing" },
    });
    fireEvent.change(screen.getByLabelText("Feedback description"), {
      target: { value: "The current crossing is unsafe." },
    });
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));

    await waitFor(() => {
      expect(createFeedbackActionMock).toHaveBeenCalledWith(
        expect.objectContaining({
          boardId: "road-repairs",
          description: "The current crossing is unsafe.",
        }),
      );
    });
  });

  it("toggles an upvote without navigating away from the feedback list", async () => {
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    fireEvent.click(screen.getAllByRole("button", { name: "Upvote" })[0]);

    const activeVote = await screen.findByRole("button", {
      name: "Remove upvote",
    });
    expect(activeVote).toHaveTextContent("13");
    expect(toggleFeedbackVoteActionMock).toHaveBeenCalledWith(
      expect.objectContaining({ itemId: "req-1", vote: 1 }),
    );
  });

  it("supports downvoting from the feedback detail page", async () => {
    render(
      <PublicPortalRequestDetailPage
        portal={publicPortalFixture}
        request={publicPortalFixture.requests[0]}
        viewer={portalViewer}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Downvote" }));

    expect(
      await screen.findByRole("button", { name: "Remove downvote" }),
    ).toBeInTheDocument();
    expect(toggleFeedbackVoteActionMock).toHaveBeenCalledWith(
      expect.objectContaining({ itemId: "req-1", vote: -1 }),
    );
  });

  it("renders signed-in portal navigation controls", () => {
    render(
      <PublicPortalRequestsPage
        portal={publicPortalFixture}
        viewer={portalViewer}
      />,
    );

    expect(
      screen.queryByRole("link", { name: "Login/signup" }),
    ).not.toBeInTheDocument();
    expect(screen.getByLabelText("Open account menu")).toBeInTheDocument();
    expect(screen.getByText("ada@example.com")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /open app/i })).toHaveAttribute(
      "href",
      "/city-roads/my-work",
    );
    expect(
      screen.getByRole("link", { name: /account settings/i }),
    ).toHaveAttribute("href", "/city-roads/settings/account");
  });

  it("copies the public portal link and confirms it with a toast", async () => {
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    fireEvent.click(screen.getByRole("button", { name: "Copy link" }));

    await waitFor(() => {
      expect(clipboardWriteTextMock).toHaveBeenCalledWith(window.location.href);
    });
    expect(toast.success).toHaveBeenCalledWith("Link copied to clipboard");
  });

  it("opens the native share sheet for the public portal", async () => {
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    fireEvent.click(screen.getByRole("button", { name: "Share" }));

    await waitFor(() => {
      expect(shareMock).toHaveBeenCalledWith({
        title: publicPortalFixture.name,
        text:
          publicPortalFixture.description ||
          `${publicPortalFixture.workspace.name} feedback`,
        url: window.location.href,
      });
    });
  });

  it("renders roadmap columns from promoted public requests", async () => {
    render(<PublicPortalRoadmapPage portal={publicPortalFixture} />);

    expect(screen.getByText("Planned")).toBeInTheDocument();
    expect(screen.getByText("In Progress")).toBeInTheDocument();
    expect(screen.getByText("Done")).toBeInTheDocument();
    expect(
      await screen.findByText("Resurface Market Road before rainy season"),
    ).toBeInTheDocument();
    expect(
      screen.queryByText("Add pedestrian crossing near East Avenue school"),
    ).not.toBeInTheDocument();
  });

  it("renders a public request detail page with comments and metadata", () => {
    render(
      <PublicPortalRequestDetailPage
        portal={publicPortalFixture}
        request={publicPortalFixture.requests[0]}
        viewer={portalViewer}
      />,
    );

    expect(
      screen.getByRole("heading", {
        name: "Add pedestrian crossing near East Avenue school",
      }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Comment")).toBeInTheDocument();
    expect(screen.getByText("Road repairs")).toBeInTheDocument();
    expect(screen.getByText("Copy link")).toBeInTheDocument();
  });

  it("adds a new public feedback comment to the discussion", async () => {
    render(
      <PublicPortalRequestDetailPage
        portal={publicPortalFixture}
        request={publicPortalFixture.requests[0]}
        viewer={portalViewer}
      />,
    );

    fireEvent.change(screen.getByLabelText("Comment"), {
      target: { value: "Thanks for raising this." },
    });
    fireEvent.click(screen.getByRole("button", { name: "Comment" }));

    expect(
      await screen.findByText("Thanks for raising this."),
    ).toBeInTheDocument();
    expect(createFeedbackCommentActionMock).toHaveBeenCalledWith(
      expect.objectContaining({ itemId: "req-1" }),
    );
  });
});
