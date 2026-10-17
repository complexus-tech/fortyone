/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { toast } from "sonner";
import { useCopyToClipboard } from "@/hooks/clipboard";
import type { DetailedStory } from "@/shared/story/types";
import { useNativeLink } from "./native-query";
import { NativeStoryFields, NativeObjectiveFields } from "./native-fields";
import { NativeLinkPreview } from "./native-preview";
import type { NativeLinkTarget } from "./native-target";
import { NativeLinkNavigationProvider } from "./native-navigation";
import { useNativePreviewTerminology } from "./native-terminology";

jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));
jest.mock("@/hooks/clipboard", () => ({ useCopyToClipboard: jest.fn() }));
jest.mock("./native-terminology", () => ({
  useNativePreviewTerminology: jest.fn(),
}));

jest.mock("./native-query", () => ({ useNativeLink: jest.fn() }));
jest.mock("./native-fields", () => ({
  NativeStoryFields: jest.fn(() => <div>Editable story fields</div>),
  NativeObjectiveFields: jest.fn(() => <div>Editable objective fields</div>),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: "acme",
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));

const target: NativeLinkTarget = {
  kind: "story",
  id: "ENG-42",
  workspaceSlug: "acme",
  href: "https://acme.fortyone.app/work/ENG-42#comments",
};
const story = {
  id: "story",
  title: "Release onboarding",
  sequenceId: 42,
  teamCode: "ENG",
} as DetailedStory;
beforeEach(() => {
  jest
    .mocked(useNativePreviewTerminology)
    .mockReturnValue({ label: "Story", term: "story" });
  jest
    .mocked(useCopyToClipboard)
    .mockReturnValue([null, jest.fn().mockResolvedValue(true)]);
  jest.mocked(useNativeLink).mockReturnValue({
    data: { kind: "story", value: story },
    loading: false,
    unavailable: false,
    retry: jest.fn(),
    refresh: jest.fn().mockResolvedValue(undefined),
    refreshing: false,
    refreshedAt: undefined,
    refreshError: null,
  });
});
afterEach(() => {
  cleanup();
  jest.clearAllMocks();
});

describe("native card variants", () => {
  it.each([
    ["Task", "task"],
    ["Issue", "issue"],
    ["Project", "project"],
    ["Goal", "goal"],
  ])(
    "uses %s terminology consistently across full, compact, and unavailable cards",
    (label, term) => {
      jest.mocked(useNativePreviewTerminology).mockReturnValue({ label, term });
      const { rerender } = render(<NativeLinkPreview target={target} />);
      expect(screen.getByText(`${label} · ENG-42`)).toBeVisible();
      expect(screen.getByRole("link", { name: `View ${term}` })).toBeVisible();
      expect(
        screen.getByRole("button", { name: `Copy ${term} link` }),
      ).toBeVisible();
      expect(screen.getByText(`Open ${term}`)).toBeInTheDocument();
      rerender(<NativeLinkPreview compact target={target} />);
      expect(screen.getByText(`${label} · ENG-42`)).toBeVisible();
      expect(
        screen.queryByRole("button", { name: /Copy/ }),
      ).not.toBeInTheDocument();
      jest.mocked(useNativeLink).mockReturnValue({
        data: null,
        loading: false,
        unavailable: true,
        retry: jest.fn(),
        refresh: jest.fn().mockResolvedValue(undefined),
        refreshing: false,
        refreshedAt: undefined,
        refreshError: null,
      });
      rerender(<NativeLinkPreview target={target} />);
      expect(screen.getByRole("link", { name: `Open ${term}` })).toBeVisible();
    },
  );

  it("copies the original deep link without opening the card and reports success", async () => {
    const copy = jest.fn().mockResolvedValue(true);
    const onOpen = jest.fn(() => true);
    jest.mocked(useCopyToClipboard).mockReturnValue([null, copy]);
    render(
      <NativeLinkNavigationProvider onOpen={onOpen}>
        <NativeLinkPreview target={target} />
      </NativeLinkNavigationProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Copy story link" }));
    expect(copy).toHaveBeenCalledWith(target.href);
    expect(onOpen).not.toHaveBeenCalled();
    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("Story link copied");
    });
  });

  it("reports clipboard failures without claiming the link was copied", async () => {
    jest
      .mocked(useCopyToClipboard)
      .mockReturnValue([null, jest.fn().mockResolvedValue(false)]);
    render(<NativeLinkPreview target={target} />);
    fireEvent.click(screen.getByRole("button", { name: "Copy story link" }));
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        "Could not copy link. Please try again.",
      );
    });
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("opens the whole card through the surrounding chat panel without losing modified-link behavior", () => {
    const onOpen = jest.fn(() => true);
    render(
      <NativeLinkNavigationProvider onOpen={onOpen}>
        <NativeLinkPreview target={target} />
      </NativeLinkNavigationProvider>,
    );
    const link = screen.getByRole("link", { name: `Open ${story.title}` });
    fireEvent.click(link);
    expect(onOpen).toHaveBeenCalledWith(target, link);
    fireEvent.click(screen.getByRole("link", { name: "View story" }));
    expect(onOpen).toHaveBeenCalledTimes(2);
    onOpen.mockClear();
    fireEvent.click(link, { metaKey: true });
    expect(onOpen).not.toHaveBeenCalled();
  });

  it("renders compact title and reference without mounting field queries, and opens locally", () => {
    render(<NativeLinkPreview compact enabled={false} target={target} />);
    expect(screen.getByText(story.title)).toBeVisible();
    expect(screen.getByText("Story · ENG-42")).toBeVisible();
    expect(NativeStoryFields).not.toHaveBeenCalled();
    expect(NativeObjectiveFields).not.toHaveBeenCalled();
    expect(useNativeLink).toHaveBeenCalledWith(
      target,
      false,
      expect.any(Boolean),
    );
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "/acme/work/ENG-42#comments",
    );
    expect(screen.getByRole("link")).not.toHaveAttribute("target");
  });

  it("mounts editable fields only on a full card in the current workspace", () => {
    const { rerender } = render(<NativeLinkPreview target={target} />);
    expect(screen.getByText("Editable story fields")).toBeVisible();
    rerender(
      <NativeLinkPreview
        target={{
          ...target,
          workspaceSlug: "other",
          href: "https://other.fortyone.app/work/ENG-42",
        }}
      />,
    );
    expect(screen.queryByText("Editable story fields")).not.toBeInTheDocument();
    expect(screen.getByText(story.title)).toBeVisible();
    expect(screen.getByRole("link", { name: "View story" })).toHaveAttribute(
      "target",
      "_blank",
    );
    expect(
      screen.getByText("Open this workspace to make changes."),
    ).toBeVisible();
  });

  it("uses a generic unavailable state without cached private titles or fields", () => {
    jest.mocked(useNativeLink).mockReturnValue({
      data: null,
      loading: false,
      unavailable: true,
      retry: jest.fn(),
      refresh: jest.fn().mockResolvedValue(undefined),
      refreshing: false,
      refreshedAt: undefined,
      refreshError: null,
    });
    render(<NativeLinkPreview target={target} />);
    expect(screen.getByText("Preview unavailable")).toBeVisible();
    expect(screen.queryByText(story.title)).not.toBeInTheDocument();
    expect(NativeStoryFields).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Retry" })).toBeVisible();
  });
});
