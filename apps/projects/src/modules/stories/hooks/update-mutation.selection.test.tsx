import type { Dispatch, SetStateAction } from "react";
import { useState } from "react";
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
} from "@tanstack/react-query";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import type { GroupedStoriesResponse, Story, StoryGroup } from "../types";
import { storyKeys } from "../constants";
import { bulkUpdateAction } from "../actions/bulk-update-stories";
import { useGroupStoriesInfinite } from "./use-group-stories-infinite";
import { useBulkUpdateStoriesMutation } from "./update-mutation";

jest.mock("next/navigation", () => ({ useParams: () => ({ teamId: "team" }) }));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace" }),
  useAnalytics: () => ({ analytics: { track: jest.fn() } }),
  useTerminology: () => ({ getTermDisplay: () => "stories" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "user" } } }),
}));
jest.mock("../actions/bulk-update-stories", () => ({
  bulkUpdateAction: jest.fn(),
}));
jest.mock("../queries/get-group-stories", () => ({
  getGroupStories: () =>
    new Promise(() => {
      /* Keep refetches pending during assertions. */
    }),
}));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), warning: jest.fn() },
}));

const createStory = (id: string, subStories: Story[] = []): Story =>
  ({ id, statusId: "started", title: id, subStories }) as Story;

const initialStories: GroupedStoriesResponse = {
  groups: [
    {
      key: "started",
      stories: [
        createStory("parent", [
          createStory("child-one"),
          createStory("child-two"),
        ]),
        createStory("sibling"),
      ],
      totalCount: 2,
      loadedCount: 2,
      hasMore: false,
      nextPage: 1,
    },
    {
      key: "backlog",
      stories: [],
      totalCount: 0,
      loadedCount: 0,
      hasMore: false,
      nextPage: 1,
    },
  ],
  meta: {
    groupBy: "status",
    filters: {},
    orderBy: "created",
    orderDirection: "desc",
    totalGroups: 2,
  },
};

type SelectionProps = {
  selected: string[];
  setSelected: Dispatch<SetStateAction<string[]>>;
};

const SelectableStory = ({
  story,
  selected,
  setSelected,
}: SelectionProps & { story: Story }) => (
  <div>
    <label>
      <input
        checked={selected.includes(story.id)}
        onChange={(event) => {
          const checked = event.target.checked;
          setSelected((previous) =>
            checked
              ? [...previous, story.id]
              : previous.filter((id) => id !== story.id),
          );
        }}
        type="checkbox"
      />
      {story.title}
    </label>
    <span data-testid={`${story.id}-status`}>{story.statusId}</span>
    {story.subStories.map((child) => (
      <SelectableStory
        key={child.id}
        selected={selected}
        setSelected={setSelected}
        story={child}
      />
    ))}
  </div>
);

const StoryGroupRows = ({
  group,
  selected,
  setSelected,
}: SelectionProps & { group: StoryGroup }) => {
  const { data } = useGroupStoriesInfinite(
    {
      groupKey: group.key,
      groupBy: "status",
      orderBy: "created",
      orderDirection: "desc",
    },
    group,
  );

  return (
    <section aria-label={group.key} hidden={group.loadedCount === 0}>
      <span data-testid={`${group.key}-count`}>{group.totalCount}</span>
      {data.pages
        .flatMap((page) => page.stories)
        .map((story) => (
          <SelectableStory
            key={story.id}
            selected={selected}
            setSelected={setSelected}
            story={story}
          />
        ))}
    </section>
  );
};

const SelectionList = () => {
  const [selected, setSelected] = useState<string[]>([]);
  const mutation = useBulkUpdateStoriesMutation();
  const { data } = useQuery({
    queryKey: storyKeys.teamGrouped("workspace", "team", { groupBy: "status" }),
    initialData: initialStories,
    staleTime: Infinity,
    queryFn: () =>
      new Promise<GroupedStoriesResponse>(() => {
        /* Keep refetches pending during assertions. */
      }),
  });

  return (
    <>
      <button
        onClick={() => {
          mutation.mutate({
            storyIds: selected,
            payload: { statusId: "backlog" },
          });
        }}
        type="button"
      >
        Move selected to backlog
      </button>
      <span data-testid="save-state">{mutation.status}</span>
      {data.groups.map((group) => (
        <StoryGroupRows
          group={group}
          key={group.key}
          selected={selected}
          setSelected={setSelected}
        />
      ))}
    </>
  );
};

describe("bulk selection list updates", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it.each([
    ["stories", ["parent", "sibling"], "success"],
    ["stories", ["parent", "sibling"], "error"],
    ["sub-stories", ["child-one", "child-two"], "success"],
    ["sub-stories", ["child-one", "child-two"], "error"],
  ] as const)(
    "updates selected %s (%s) optimistically through %s",
    async (kind, selectedIds, outcome) => {
      const client = new QueryClient({
        defaultOptions: {
          queries: { retry: false },
          mutations: { retry: false },
        },
      });
      let finishUpdate:
        | ((response: Awaited<ReturnType<typeof bulkUpdateAction>>) => void)
        | undefined;
      jest.mocked(bulkUpdateAction).mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finishUpdate = resolve;
          }),
      );
      const { unmount } = render(
        <QueryClientProvider client={client}>
          <SelectionList />
        </QueryClientProvider>,
      );

      for (const id of selectedIds)
        fireEvent.click(screen.getByRole("checkbox", { name: id }));
      fireEvent.click(
        screen.getByRole("button", { name: "Move selected to backlog" }),
      );
      await waitFor(() => {
        expect(bulkUpdateAction).toHaveBeenCalledTimes(1);
      });

      expect(screen.getByTestId("save-state")).toHaveTextContent("pending");
      for (const id of selectedIds) {
        expect(screen.getByTestId(`${id}-status`)).toHaveTextContent("backlog");
      }
      if (kind === "stories") {
        const backlog = screen.getByRole("region", { name: "backlog" });
        expect(backlog).toBeVisible();
        expect(
          within(backlog).getByRole("checkbox", { name: "parent" }),
        ).toBeChecked();
        expect(screen.getByTestId("backlog-count")).toHaveTextContent("2");
        expect(
          screen.queryByRole("region", { name: "started" }),
        ).not.toBeInTheDocument();
        expect(screen.getByTestId("child-one-status")).toHaveTextContent(
          "started",
        );
      } else {
        expect(screen.getByRole("region", { name: "started" })).toBeVisible();
        expect(
          screen.queryByRole("region", { name: "backlog" }),
        ).not.toBeInTheDocument();
        expect(screen.getByTestId("parent-status")).toHaveTextContent(
          "started",
        );
        expect(screen.getByTestId("started-count")).toHaveTextContent("2");
      }

      act(() => {
        finishUpdate?.(
          outcome === "success"
            ? {
                data: {
                  totalCount: 2,
                  succeededCount: 2,
                  failedCount: 0,
                  partial: false,
                  items: selectedIds.map((storyId) => ({
                    storyId,
                    success: true,
                  })),
                },
              }
            : { data: null, error: { message: "Status update rejected" } },
        );
      });
      await waitFor(() => {
        expect(screen.getByTestId("save-state")).toHaveTextContent(outcome);
      });

      for (const id of selectedIds) {
        expect(screen.getByTestId(`${id}-status`)).toHaveTextContent(
          outcome === "success" ? "backlog" : "started",
        );
      }
      if (outcome === "error") {
        expect(screen.getByRole("region", { name: "started" })).toBeVisible();
        expect(screen.getByTestId("started-count")).toHaveTextContent("2");
        expect(
          screen.queryByRole("region", { name: "backlog" }),
        ).not.toBeInTheDocument();
      }
      unmount();
      client.clear();
    },
  );
});
