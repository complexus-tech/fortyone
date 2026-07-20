/* global beforeAll, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactTypes from "react";
import {
  act,
  fireEvent,
  render as renderWithTestingLibrary,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import { publicPortalFixture } from "./fixtures";
import {
  createFeedbackAction,
  createFeedbackCommentAction,
  toggleFeedbackVoteAction,
} from "./actions";
import {
  getPublicPortalNotificationsAction,
  getPublicPortalUnreadCountAction,
  markPublicPortalNotificationReadAction,
} from "./notification-actions";
import {
  PublicPortalAuthorProfilePage,
  PublicPortalRequestDetailPage,
  PublicPortalRequestsPage,
  PublicPortalRoadmapPage,
} from ".";

const clipboardWriteTextMock = jest.fn(async (_text: string) => undefined);
const shareMock = jest.fn(async (_data?: ShareData) => undefined);
let triggerIntersection: (() => void) | undefined;

const createDeferred = <T,>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });

  return { promise, resolve };
};

const portalViewer = {
  accountHref: "/portal/city-roads/account",
  appHref: "/city-roads/my-work",
  avatarUrl: null,
  email: "ada@example.com",
  feedbackSetupHref: "/city-roads/settings/workspace/feedback",
  id: "00000000-0000-4000-8000-000000000001",
  name: "Ada Ndlovu",
};

const createRoadmapQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        gcTime: Infinity,
        retry: false,
      },
    },
  });

const renderRoadmap = (
  queryClient = createRoadmapQueryClient(),
  portal = publicPortalFixture,
) =>
  renderWithTestingLibrary(
    <QueryClientProvider client={queryClient}>
      <PublicPortalRoadmapPage portal={portal} />
    </QueryClientProvider>,
  );

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
    ExternalLinkIcon: Icon,
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

jest.mock("@/utils", () => ({
  buildWorkspaceUrl: (slug: string, path = "/my-work") => `/${slug}${path}`,
  hexToRgba: (color: string, opacity: number) =>
    `color-mix(in srgb, ${color} ${opacity * 100}%, transparent)`,
}));

jest.mock("./actions", () => ({
  createFeedbackAction: jest.fn(),
  createFeedbackCommentAction: jest.fn(),
  toggleFeedbackVoteAction: jest.fn(),
}));

jest.mock("./notification-actions", () => ({
  getPublicPortalNotificationsAction: jest.fn(async () => ({
    data: {
      notifications: [],
      pagination: { hasMore: false, nextPage: 2, page: 1, pageSize: 20 },
    },
  })),
  getPublicPortalUnreadCountAction: jest.fn(async () => ({
    data: { count: 0 },
  })),
  markPublicPortalNotificationReadAction: jest.fn(async () => ({
    data: null,
  })),
}));

const createFeedbackActionMock = jest.mocked(createFeedbackAction);
const createFeedbackCommentActionMock = jest.mocked(
  createFeedbackCommentAction,
);
const toggleFeedbackVoteActionMock = jest.mocked(toggleFeedbackVoteAction);
const getPublicPortalNotificationsActionMock = jest.mocked(
  getPublicPortalNotificationsAction,
);
const getPublicPortalUnreadCountActionMock = jest.mocked(
  getPublicPortalUnreadCountAction,
);
const markPublicPortalNotificationReadActionMock = jest.mocked(
  markPublicPortalNotificationReadAction,
);

const render = (element: ReactTypes.ReactElement) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      mutations: { retry: false },
      queries: { retry: false },
    },
  });

  return renderWithTestingLibrary(
    <QueryClientProvider client={queryClient}>{element}</QueryClientProvider>,
  );
};

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
}));

jest.mock("ui", () => {
  const React: typeof ReactTypes = jest.requireActual("react");

  const Box = ({
    as: Tag = "div",
    children,
    ...props
  }: ReactTypes.HTMLAttributes<HTMLElement> &
    Record<string, unknown> & {
      as?: ReactTypes.ElementType;
    }) => React.createElement(Tag, getDomProps(props), children);
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
  const PopoverContext = React.createContext<{
    onOpenChange?: (open: boolean) => void;
    open: boolean;
  }>({ open: false });
  const Popover = ({
    children,
    onOpenChange,
    open = false,
  }: {
    children: ReactTypes.ReactNode;
    onOpenChange?: (open: boolean) => void;
    open?: boolean;
  }) => (
    <PopoverContext.Provider value={{ onOpenChange, open }}>
      <div>{children}</div>
    </PopoverContext.Provider>
  );
  const PopoverTrigger = ({ children }: { children: ReactTypes.ReactNode }) => (
    <PopoverContext.Consumer>
      {({ onOpenChange, open }) => {
        if (!React.isValidElement(children)) return children;

        const trigger = children as ReactTypes.ReactElement<{
          onClick?: ReactTypes.MouseEventHandler;
        }>;
        return React.cloneElement(trigger, {
          onClick: (event: ReactTypes.MouseEvent) => {
            trigger.props.onClick?.(event);
            onOpenChange?.(!open);
          },
        });
      }}
    </PopoverContext.Consumer>
  );
  const PopoverContent = ({ children }: { children: ReactTypes.ReactNode }) => (
    <PopoverContext.Consumer>
      {({ open }) => (open ? <div>{children}</div> : null)}
    </PopoverContext.Consumer>
  );
  Popover.Trigger = PopoverTrigger;
  Popover.Content = PopoverContent;
  const TimeAgo = ({ timestamp }: { timestamp: string }) => (
    <span>{timestamp}</span>
  );
  const TabsContext = React.createContext<{
    onValueChange?: (value: string) => void;
    value: string;
  }>({ value: "" });
  const Tabs = ({
    children,
    defaultValue = "",
    onValueChange,
    value = defaultValue,
    ...props
  }: ReactTypes.HTMLAttributes<HTMLDivElement> & {
    defaultValue?: string;
    onValueChange?: (value: string) => void;
    value?: string;
  }) => (
    <TabsContext.Provider value={{ onValueChange, value }}>
      <div {...getDomProps(props)}>{children}</div>
    </TabsContext.Provider>
  );
  const TabsList = ({
    children,
    ...props
  }: ReactTypes.HTMLAttributes<HTMLDivElement>) => (
    <div role="tablist" {...props}>
      {children}
    </div>
  );
  const TabsTab = ({
    children,
    value,
    ...props
  }: ReactTypes.ButtonHTMLAttributes<HTMLButtonElement> & {
    value: string;
  }) => (
    <TabsContext.Consumer>
      {(context) => (
        <button
          aria-selected={context.value === value}
          onClick={() => {
            context.onValueChange?.(value);
          }}
          role="tab"
          type="button"
          {...props}
        >
          {children}
        </button>
      )}
    </TabsContext.Consumer>
  );
  const TabsPanel = ({
    children,
    value,
    ...props
  }: ReactTypes.HTMLAttributes<HTMLDivElement> & { value: string }) => (
    <TabsContext.Consumer>
      {(context) =>
        context.value === value ? <div {...props}>{children}</div> : null
      }
    </TabsContext.Consumer>
  );
  Tabs.List = TabsList;
  Tabs.Tab = TabsTab;
  Tabs.Panel = TabsPanel;

  return {
    Avatar,
    Box,
    Button,
    Dialog,
    Flex,
    Input,
    Menu,
    Popover,
    Text,
    TextArea,
    TextEditor,
    TimeAgo,
    Tabs,
  };
});

describe("Public portal UI", () => {
  beforeAll(() => {
    class IntersectionObserverMock {
      constructor(callback: IntersectionObserverCallback) {
        triggerIntersection = () => {
          callback(
            [{ isIntersecting: true } as IntersectionObserverEntry],
            this as unknown as IntersectionObserver,
          );
        };
      }

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
    triggerIntersection = undefined;
    window.history.replaceState({}, "", "/portal/city-roads/feedback");
    createFeedbackActionMock.mockImplementation(async (input) => ({
      data: {
        ...publicPortalFixture.requests[0],
        id: "feedback-new",
        slug: "new-feedback",
        boardId: input.boardId,
        description: input.description,
        title: input.title,
      },
    }));
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
    getPublicPortalNotificationsActionMock.mockResolvedValue({
      data: {
        notifications: [],
        pagination: { hasMore: false, nextPage: 2, page: 1, pageSize: 20 },
      },
    });
    getPublicPortalUnreadCountActionMock.mockResolvedValue({
      data: { count: 0 },
    });
    markPublicPortalNotificationReadActionMock.mockResolvedValue({
      data: null,
    });
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
        const search = requestUrl.searchParams.get("search")?.toLowerCase();
        const sort = requestUrl.searchParams.get("sort");
        const page = requestUrl.searchParams.get("page");
        if (page === "2") {
          return {
            json: async () => ({
              data: {
                ...publicPortalFixture,
                requests: [
                  {
                    ...publicPortalFixture.requests[0],
                    id: "req-5",
                    slug: "new-page-feedback",
                    title: "New page feedback",
                  },
                ],
                requestsHasMore: false,
              },
            }),
            ok: true,
          } as Response;
        }
        const requests = publicPortalFixture.requests.filter(
          (request) =>
            (!status || request.status === status) &&
            (!boardId || request.boardId === boardId) &&
            (!search ||
              `${request.title} ${request.description}`
                .toLowerCase()
                .includes(search)),
        );

        requests.sort((first, second) => {
          if (sort === "oldest") return first.id.localeCompare(second.id);
          if (sort === "newest") return second.id.localeCompare(first.id);
          return second.voteCount - first.voteCount;
        });

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
    expect(
      screen.getByRole("link", { name: "City Roads Program feedback" }),
    ).toHaveAttribute("href", "/portal/city-roads/feedback");
    expect(
      screen.getByRole("link", { name: "Create your own board" }),
    ).toHaveAttribute("href", "/city-roads/settings/workspace/feedback");
  });

  it("hides comment metadata when feedback has no comments", () => {
    const portal = {
      ...publicPortalFixture,
      requests: [
        {
          ...publicPortalFixture.requests[0],
          commentCount: 0,
        },
      ],
    };

    render(<PublicPortalRequestsPage portal={portal} />);

    expect(screen.queryByLabelText("0 comments")).not.toBeInTheDocument();
  });

  it("sends logged-out visitors to login before submitting feedback", () => {
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    expect(
      screen.getByRole("link", { name: "Login to submit feedback" }),
    ).toHaveAttribute(
      "href",
      "/?callbackUrl=%2Fportal%2Fcity-roads%2Ffeedback",
    );
    expect(
      screen.queryByRole("button", { name: "New Feedback" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Create your own board" }),
    ).toHaveAttribute(
      "href",
      "/signup?source=portal&callbackUrl=%2Fonboarding%2Fcreate%3FcallbackUrl%3D%252Fsettings%252Fworkspace%252Ffeedback",
    );
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

  it("searches only after submitting and clears back to all feedback", async () => {
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    const search = screen.getByPlaceholderText("Search feedback...");
    fireEvent.change(search, { target: { value: "storm drain" } });

    expect(global.fetch).not.toHaveBeenCalled();

    fireEvent.submit(search.closest("form")!);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("search=storm+drain"),
      );
    });
    expect(window.location.search).toContain("search=storm+drain");
    expect(
      await screen.findByText("Blocked storm drain on 4th Street"),
    ).toBeVisible();

    const activeSearch = screen.getByPlaceholderText("Search feedback...");
    fireEvent.change(activeSearch, { target: { value: "" } });
    fireEvent.submit(activeSearch.closest("form")!);

    await waitFor(() => {
      expect(window.location.search).not.toContain("search=");
      expect(
        screen.getByText("Add pedestrian crossing near East Avenue school"),
      ).toBeVisible();
    });
  });

  it("stores board, status, and sort filters in the URL", async () => {
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    fireEvent.click(screen.getByRole("button", { name: "Drainage" }));
    await waitFor(() => {
      const params = new URLSearchParams(window.location.search);
      expect(params.get("boardId")).toBe("drainage");
      expect(params.get("sort")).toBe("top");
    });
    fireEvent.click(screen.getByRole("button", { name: "Reviewing" }));
    await waitFor(() => {
      expect(new URLSearchParams(window.location.search).get("status")).toBe(
        "reviewing",
      );
    });
    fireEvent.click(screen.getByRole("button", { name: /newest/i }));

    await waitFor(() => {
      const params = new URLSearchParams(window.location.search);
      expect(params.get("boardId")).toBe("drainage");
      expect(params.get("status")).toBe("reviewing");
      expect(params.get("sort")).toBe("newest");
      expect(
        screen.queryByText("Add pedestrian crossing near East Avenue school"),
      ).not.toBeInTheDocument();
    });
  });

  it("loads the next feedback page when the list sentinel becomes visible", async () => {
    render(
      <PublicPortalRequestsPage
        portal={{ ...publicPortalFixture, requestsHasMore: true }}
      />,
    );

    await waitFor(() => {
      expect(triggerIntersection).toBeDefined();
    });
    act(() => {
      triggerIntersection?.();
    });

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("page=2"),
      );
      expect(screen.getByText("New page feedback")).toBeVisible();
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

  it("shows new feedback immediately and rolls it back when submission fails", async () => {
    const submission =
      createDeferred<Awaited<ReturnType<typeof createFeedbackAction>>>();
    createFeedbackActionMock.mockReturnValueOnce(submission.promise);
    const portal = {
      ...publicPortalFixture,
      boards: [publicPortalFixture.boards[0]],
    };
    render(<PublicPortalRequestsPage portal={portal} viewer={portalViewer} />);

    fireEvent.change(screen.getByLabelText("Feedback title"), {
      target: { value: "Add protected bike parking" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));

    expect(
      await screen.findByText("Add protected bike parking"),
    ).toBeInTheDocument();

    await act(async () => {
      submission.resolve({
        data: null,
        error: { message: "Unable to submit feedback" },
      });
      await submission.promise;
    });

    await waitFor(() => {
      expect(
        screen.queryByText("Add protected bike parking"),
      ).not.toBeInTheDocument();
    });
    expect(toast.error).toHaveBeenCalledWith("Feedback", {
      description: "Unable to submit feedback",
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

  it("rolls an optimistic vote back when the vote request fails", async () => {
    const voteRequest =
      createDeferred<Awaited<ReturnType<typeof toggleFeedbackVoteAction>>>();
    toggleFeedbackVoteActionMock.mockReturnValueOnce(voteRequest.promise);
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    fireEvent.click(screen.getAllByRole("button", { name: "Upvote" })[0]);

    const optimisticVote = await screen.findByRole("button", {
      name: "Remove upvote",
    });
    expect(optimisticVote).toHaveTextContent("13");

    await act(async () => {
      voteRequest.resolve({
        data: null,
        error: { message: "Unable to save vote" },
      });
      await voteRequest.promise;
    });

    await waitFor(() => {
      expect(
        screen.getAllByRole("button", { name: "Upvote" })[0],
      ).toHaveTextContent("12");
    });
    expect(toast.error).toHaveBeenCalledWith("Vote", {
      description: "Unable to save vote",
    });
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
    expect(screen.getByRole("link", { name: "Open FortyOne" })).toHaveAttribute(
      "href",
      "/city-roads/my-work",
    );
    expect(
      screen.getByRole("link", { name: /account settings/i }),
    ).toHaveAttribute("href", "/portal/city-roads/account");
  });

  it("shows feedback notifications and marks an unread item as read", async () => {
    getPublicPortalNotificationsActionMock.mockResolvedValue({
      data: {
        notifications: [
          {
            id: "notification-1",
            type: "feedback_comment",
            title: "New feedback comment",
            message: {
              template: "{actor} commented on your feedback",
              variables: {
                actor: { type: "string", value: "Tariro Moyo" },
              },
            },
            actor: {
              id: "user-2",
              name: "Tariro Moyo",
              avatarUrl: null,
            },
            feedback: {
              id: "req-1",
              title: "Add pedestrian crossing near East Avenue school",
              slug: "add-pedestrian-crossing-near-east-avenue-school",
              path: "/feedback/add-pedestrian-crossing-near-east-avenue-school",
            },
            createdAt: "2026-07-20T08:00:00.000Z",
            readAt: null,
          },
        ],
        pagination: { hasMore: false, nextPage: 2, page: 1, pageSize: 20 },
      },
    });
    getPublicPortalUnreadCountActionMock.mockResolvedValue({
      data: { count: 1 },
    });

    render(
      <PublicPortalRequestsPage
        portal={publicPortalFixture}
        viewer={portalViewer}
      />,
    );

    expect(getPublicPortalNotificationsActionMock).not.toHaveBeenCalled();
    const notificationsButton = screen.getByRole("button", {
      name: "Notifications",
    });
    fireEvent.click(notificationsButton);

    const notificationMessage = await screen.findByText(
      "Tariro Moyo commented on your feedback",
    );
    expect(getPublicPortalNotificationsActionMock).toHaveBeenCalledTimes(1);
    const notificationLink = notificationMessage.closest("a");
    expect(notificationLink).toHaveAttribute(
      "href",
      "/portal/city-roads/feedback/add-pedestrian-crossing-near-east-avenue-school",
    );
    if (!notificationLink)
      throw new Error("Notification link was not rendered");
    notificationLink.addEventListener("click", (event) => {
      event.preventDefault();
    });

    fireEvent.click(notificationLink);

    await waitFor(() => {
      expect(markPublicPortalNotificationReadActionMock).toHaveBeenCalledWith({
        notificationId: "notification-1",
        portalSlug: "city-roads",
      });
    });
  });

  it("reuses fresh notification data when the popover is reopened", async () => {
    render(
      <PublicPortalRequestsPage
        portal={publicPortalFixture}
        viewer={portalViewer}
      />,
    );

    const notificationsButton = screen.getByRole("button", {
      name: "Notifications",
    });
    fireEvent.click(notificationsButton);
    expect(await screen.findByText("You're all caught up")).toBeInTheDocument();
    expect(getPublicPortalNotificationsActionMock).toHaveBeenCalledTimes(1);

    fireEvent.click(notificationsButton);
    fireEvent.click(notificationsButton);
    expect(await screen.findByText("You're all caught up")).toBeInTheDocument();
    expect(getPublicPortalNotificationsActionMock).toHaveBeenCalledTimes(1);
  });

  it("keeps portal controls but hides workspace actions for external users", () => {
    render(
      <PublicPortalRequestsPage
        portal={publicPortalFixture}
        viewer={{
          accountHref: "/portal/city-roads/account",
          avatarUrl: null,
          email: "external@example.com",
          feedbackSetupHref: "/onboarding/create",
          id: "00000000-0000-4000-8000-000000000099",
          name: "External Contributor",
        }}
      />,
    );

    expect(
      screen.getByRole("link", { name: /account settings/i }),
    ).toHaveAttribute("href", "/portal/city-roads/account");
    expect(
      screen.queryByRole("link", { name: "Open FortyOne" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Notifications" }),
    ).toBeInTheDocument();
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
        text: `${publicPortalFixture.workspace.name} feedback`,
        url: window.location.href,
      });
    });
  });

  it("renders roadmap columns from promoted public requests", async () => {
    renderRoadmap();

    expect(screen.getByText("Planned")).toBeInTheDocument();
    expect(screen.getByText("In Progress")).toBeInTheDocument();
    expect(screen.getByText("Done")).toBeInTheDocument();
    expect(screen.getByText("Committed and queued")).toBeInTheDocument();
    expect(screen.getByText("Actively being delivered")).toBeInTheDocument();
    expect(screen.getByText("Recently completed")).toBeInTheDocument();
    expect(
      await screen.findByText("Resurface Market Road before rainy season"),
    ).toBeInTheDocument();
    expect(screen.getByText("Nothing in progress")).toBeInTheDocument();
    expect(
      screen.queryByText("Add pedestrian crossing near East Avenue school"),
    ).not.toBeInTheDocument();
  });

  it("renders a Kanban-style roadmap card with an author profile link and naked vote count", async () => {
    renderRoadmap();

    const title = await screen.findByText(
      "Resurface Market Road before rainy season",
    );
    const requestLink = title.closest("a");
    const card = requestLink?.parentElement;

    expect(requestLink).toHaveAttribute(
      "href",
      "/portal/city-roads/feedback/resurface-market-road-before-rainy-season",
    );
    expect(card?.firstElementChild).toBe(requestLink);
    expect(card).toHaveTextContent("Public Works");
    expect(card).toHaveTextContent("Road repairs");
    expect(screen.getByRole("link", { name: /Public Works/ })).toHaveAttribute(
      "href",
      "/portal/city-roads/people/00000000-0000-4000-8000-000000000003",
    );

    const voteCount = screen.getByText("18");
    expect(voteCount.closest("button")).toBeNull();
  });

  it("renders an author profile and keeps the author filter on subsequent pages", async () => {
    const authorRequest = publicPortalFixture.requests[2];
    const contributor = {
      avatarUrl: authorRequest.authorAvatar,
      id: authorRequest.authorId,
      joinedAt: "2025-04-15T10:00:00.000Z",
      name: authorRequest.authorName,
      stats: {
        commentCount: 3,
        feedbackCount: 5,
        voteScore: 18,
      },
    };
    const authorPortal = {
      ...publicPortalFixture,
      requests: [authorRequest],
      requestsHasMore: true,
    };

    render(
      <PublicPortalAuthorProfilePage
        authorId={authorRequest.authorId}
        contributor={contributor}
        initialComments={{
          comments: [
            {
              body: "The crossing should include accessible signals.",
              createdAtLabel: "Yesterday",
              feedback: {
                id: publicPortalFixture.requests[0].id,
                slug: publicPortalFixture.requests[0].slug,
                title: publicPortalFixture.requests[0].title,
              },
              id: "comment-profile-1",
            },
          ],
          pagination: {
            hasMore: false,
            nextPage: 2,
            page: 1,
            pageSize: 20,
          },
        }}
        portal={authorPortal}
        viewer={portalViewer}
      />,
    );

    expect(
      screen.getByRole("heading", { name: authorRequest.authorName }),
    ).toBeInTheDocument();
    expect(screen.getByText(authorRequest.title)).toBeInTheDocument();
    expect(screen.getByText("5")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.getAllByText("18")).toHaveLength(2);
    expect(screen.getByText("Total contributions")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("tab", { name: "Comments" }));
    expect(window.location.search).toBe("?tab=comments");
    expect(
      screen.getByText("The crossing should include accessible signals."),
    ).toBeInTheDocument();
    expect(
      screen.getByText(publicPortalFixture.requests[0].title),
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole("tab", { name: "Feedback" }));
    expect(window.location.search).toBe("");

    await waitFor(() => {
      expect(triggerIntersection).toBeDefined();
    });
    act(() => {
      triggerIntersection?.();
    });

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining(`authorId=${authorRequest.authorId}`),
      );
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("page=2"),
      );
    });
  });

  it("reuses cached roadmap columns when returning to the roadmap", async () => {
    const queryClient = createRoadmapQueryClient();
    const firstRender = renderRoadmap(queryClient);

    expect(
      await screen.findByText("Resurface Market Road before rainy season"),
    ).toBeInTheDocument();
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(3);
    });

    firstRender.unmount();
    renderRoadmap(queryClient);

    expect(
      screen.getByText("Resurface Market Road before rainy season"),
    ).toBeInTheDocument();
    expect(global.fetch).toHaveBeenCalledTimes(3);
  });

  it("loads the next roadmap page when a column sentinel becomes visible", async () => {
    const plannedRequest = publicPortalFixture.requests.find(
      (request) => request.status === "planned",
    )!;
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
        const page = requestUrl.searchParams.get("page");
        const status = requestUrl.searchParams.get("status");
        const isPlanned = status === "planned";
        let requests = isPlanned ? [plannedRequest] : [];

        if (isPlanned && page === "2") {
          requests = [
            {
              ...plannedRequest,
              id: "req-roadmap-page-2",
              slug: "roadmap-page-two",
              title: "Roadmap page two",
            },
          ];
        }

        return {
          json: async () => ({
            data: {
              ...publicPortalFixture,
              requests,
              requestsHasMore: isPlanned && page === "1",
            },
          }),
          ok: true,
        } as Response;
      },
    );

    renderRoadmap();

    expect(
      await screen.findByText("Resurface Market Road before rainy season"),
    ).toBeInTheDocument();
    await waitFor(() => {
      expect(triggerIntersection).toBeDefined();
    });
    act(() => {
      triggerIntersection?.();
    });

    expect(await screen.findByText("Roadmap page two")).toBeInTheDocument();
    expect(global.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/page=2.*status=planned/),
    );
  });

  it("renders a roadmap empty state when no feedback has been planned", async () => {
    global.fetch = jest.fn(
      async () =>
        ({
          json: async () => ({
            data: {
              ...publicPortalFixture,
              requests: [],
              requestsHasMore: false,
            },
          }),
          ok: true,
        }) as Response,
    );

    renderRoadmap();

    expect(
      await screen.findByRole("heading", {
        name: "Nothing is on the roadmap yet",
      }),
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        "Feedback will appear here once the team plans it for delivery.",
      ),
    ).toBeInTheDocument();
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

  it("returns logged-out commenters to the same feedback item", () => {
    render(
      <PublicPortalRequestDetailPage
        portal={publicPortalFixture}
        request={publicPortalFixture.requests[0]}
      />,
    );

    const loginLinks = screen.getAllByRole("link", { name: "Login/signup" });
    expect(loginLinks).toHaveLength(2);
    loginLinks.forEach((link) => {
      expect(link).toHaveAttribute(
        "href",
        "/?callbackUrl=%2Fportal%2Fcity-roads%2Ffeedback%2Fadd-pedestrian-crossing-near-east-avenue-school",
      );
    });
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

  it("rolls an optimistic comment back and restores the draft on failure", async () => {
    const commentRequest =
      createDeferred<Awaited<ReturnType<typeof createFeedbackCommentAction>>>();
    createFeedbackCommentActionMock.mockReturnValueOnce(commentRequest.promise);
    render(
      <PublicPortalRequestDetailPage
        portal={publicPortalFixture}
        request={publicPortalFixture.requests[0]}
        viewer={portalViewer}
      />,
    );

    fireEvent.change(screen.getByLabelText("Comment"), {
      target: { value: "This should appear immediately." },
    });
    fireEvent.click(screen.getByRole("button", { name: "Comment" }));

    expect(
      await screen.findByText("This should appear immediately."),
    ).toBeInTheDocument();
    expect(screen.getByText("3 comments")).toBeInTheDocument();

    await act(async () => {
      commentRequest.resolve({
        data: null,
        error: { message: "Unable to add comment" },
      });
      await commentRequest.promise;
    });

    await waitFor(() => {
      expect(
        screen.queryByText("This should appear immediately."),
      ).not.toBeInTheDocument();
      expect(screen.getByText("2 comments")).toBeInTheDocument();
    });
    expect(screen.getByLabelText("Comment")).toHaveValue(
      "This should appear immediately.",
    );
    expect(toast.error).toHaveBeenCalledWith("Comment", {
      description: "Unable to add comment",
    });
  });

  it("preserves a newer comment draft when an earlier submission fails", async () => {
    const commentRequest =
      createDeferred<Awaited<ReturnType<typeof createFeedbackCommentAction>>>();
    createFeedbackCommentActionMock.mockReturnValueOnce(commentRequest.promise);
    render(
      <PublicPortalRequestDetailPage
        portal={publicPortalFixture}
        request={publicPortalFixture.requests[0]}
        viewer={portalViewer}
      />,
    );

    const commentEditor = screen.getByLabelText("Comment");
    fireEvent.change(commentEditor, {
      target: { value: "The failed comment." },
    });
    fireEvent.click(screen.getByRole("button", { name: "Comment" }));
    fireEvent.change(commentEditor, {
      target: { value: "A newer draft written while saving." },
    });

    await act(async () => {
      commentRequest.resolve({
        data: null,
        error: { message: "Unable to add comment" },
      });
      await commentRequest.promise;
    });

    const commentButton = screen.getByRole("button", { name: "Comment" });
    await waitFor(() => {
      expect(commentButton).toBeEnabled();
    });
    fireEvent.click(commentButton);

    await waitFor(() => {
      expect(createFeedbackCommentActionMock).toHaveBeenLastCalledWith(
        expect.objectContaining({
          body: "A newer draft written while saving.",
        }),
      );
    });
  });
});
