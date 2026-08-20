/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { StoryGitHubLink } from "@/modules/settings/workspace/integrations/github/types";
import type { StoryFeedbackLink } from "@/modules/team-feedback/types";
import type { DetailedStory } from "../types";
import { StoryBanners } from "./story-banners";

const mockGitHubLinks: StoryGitHubLink[] = [];
const mockFeedbackLinks: StoryFeedbackLink[] = [];
const mockRequestLinks: unknown[] = [];

jest.mock("@/hooks", () => ({
  useTerminology: () => ({
    getTermDisplay: () => "ticket",
  }),
}));

jest.mock("@/lib/hooks/github", () => ({
  useStoryGitHubLinks: () => ({ data: mockGitHubLinks }),
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
});
