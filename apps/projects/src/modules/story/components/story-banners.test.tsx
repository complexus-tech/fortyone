/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { StoryGitHubLink } from "@/modules/settings/workspace/integrations/github/types";
import type { StoryFeedbackLink } from "@/modules/team-feedback/types";
import type { DetailedStory } from "../types";
import { StoryBanners } from "./story-banners";

const mockGitHubLinks: StoryGitHubLink[] = [];
const mockFeedbackLinks: StoryFeedbackLink[] = [];
const mockRequestLinks: unknown[] = [];
const mockRetryMutate = jest.fn();
const mockOverrideMutate = jest.fn();

jest.mock("@/hooks", () => ({
  useTerminology: () => ({
    getTermDisplay: () => "ticket",
  }),
}));

jest.mock("@/lib/hooks/github", () => ({
  useStoryGitHubLinks: () => ({ data: mockGitHubLinks }),
}));

jest.mock("@/lib/hooks/calendar/use-schedule-issues", () => ({
  useOverrideCalendarScheduleIssue: () => ({
    isPending: false,
    mutate: mockOverrideMutate,
  }),
  useRetryCalendarScheduleIssue: () => ({
    isPending: false,
    mutate: mockRetryMutate,
    variables: undefined,
  }),
}));

jest.mock("@/components/ui/story/schedule-issue-dialog", () => ({
  ScheduleIssueDialog: ({
    issue,
    onSubmit,
  }: {
    issue: { storyCode: string };
    onSubmit: (startAt: string) => void;
  }) => (
    <div role="dialog">
      Choose a time for {issue.storyCode}
      <button
        onClick={() => {
          onSubmit("2026-08-21T08:00:00.000Z");
        }}
        type="button"
      >
        Confirm time
      </button>
    </div>
  ),
}));

jest.mock("@/modules/team-feedback/hooks/use-story-feedback-links", () => ({
  useStoryFeedbackLinks: () => ({ data: mockFeedbackLinks }),
}));

jest.mock(
  "@/modules/integration-requests/hooks/use-story-request-links",
  () => ({
    useStoryIntegrationRequestLinks: () => ({ data: mockRequestLinks }),
  }),
);

jest.mock("./github-banner", () => ({
  GitHubBannerRow: ({ link }: { link: StoryGitHubLink }) => (
    <div data-testid="github-banner">{link.title}</div>
  ),
}));

jest.mock("./feedback-banner", () => ({
  FeedbackBannerRow: ({ link }: { link: StoryFeedbackLink }) => (
    <div data-testid="feedback-banner">{link.feedbackTitle}</div>
  ),
}));

jest.mock("./integration-request-banner", () => ({
  IntegrationRequestBannerRow: () => <div data-testid="request-banner" />,
}));

const createStory = (overrides: Partial<DetailedStory> = {}): DetailedStory =>
  ({
    assigneeId: "user-1",
    autoSchedulingEnabled: true,
    autoSchedulingLocked: false,
    autoSchedulingReason: null,
    autoSchedulingStatus: "needs_time",
    estimatedDurationMinutes: null,
    id: "story-1",
    ...overrides,
  }) as DetailedStory;

describe("StoryBanners", () => {
  beforeEach(() => {
    mockGitHubLinks.length = 0;
    mockFeedbackLinks.length = 0;
    mockRequestLinks.length = 0;
    mockRetryMutate.mockReset();
    mockOverrideMutate.mockReset();
  });

  it("shows Maya's actionable scheduling reason with workspace terminology", () => {
    render(<StoryBanners story={createStory()} />);

    expect(
      screen.getByTitle(/Maya needs your help.*Add time needed/i),
    ).toBeInTheDocument();
  });

  it("stacks Maya and linked-source banners behind one paginated surface", () => {
    mockGitHubLinks.push({
      externalType: "issue",
      id: "github-1",
      title: "GitHub issue",
    } as StoryGitHubLink);
    mockFeedbackLinks.push({
      feedbackTitle: "Customer feedback",
      id: "feedback-1",
    } as StoryFeedbackLink);

    render(
      <StoryBanners
        story={createStory({
          autoSchedulingReason: "Maya could not fit this story this week.",
          autoSchedulingStatus: "cannot_fit",
          estimatedDurationMinutes: 60,
        })}
      />,
    );

    expect(screen.getByText("1 of 3")).toBeInTheDocument();
    expect(
      screen.getByTitle(/Maya could not fit this ticket this week/i),
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Next story banner" }));
    expect(screen.getByTestId("github-banner")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Next story banner" }));
    expect(screen.getByTestId("feedback-banner")).toBeInTheDocument();
  });

  it("offers retry and manual placement for a story Maya cannot fit", () => {
    render(
      <StoryBanners
        story={createStory({
          autoSchedulingStatus: "cannot_fit",
          estimatedDurationMinutes: 60,
          sequenceId: 42,
          teamCode: "ENG",
          title: "Ship banner actions",
        })}
      />,
    );

    fireEvent.click(
      screen.getByRole("button", { name: "More Maya scheduling actions" }),
    );
    fireEvent.click(screen.getByRole("menuitem", { name: "Retry" }));
    expect(mockRetryMutate).toHaveBeenCalledWith("story-1");

    fireEvent.click(
      screen.getByRole("button", { name: "More Maya scheduling actions" }),
    );
    fireEvent.click(screen.getByRole("menuitem", { name: "Choose time" }));
    expect(screen.getByRole("dialog")).toHaveTextContent(
      "Choose a time for ENG-42",
    );

    fireEvent.click(screen.getByRole("button", { name: "Confirm time" }));
    expect(mockOverrideMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        startAt: "2026-08-21T08:00:00.000Z",
        storyId: "story-1",
      }),
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
  });

  it("does not show scheduling actions before Maya has a placement failure", () => {
    render(<StoryBanners story={createStory()} />);

    expect(
      screen.queryByRole("button", { name: "More Maya scheduling actions" }),
    ).not.toBeInTheDocument();
  });
});
