import type { ComponentProps } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import type { KeyResult } from "@/modules/objectives/types";
import { KeyResultContextMenu } from "./key-result-context-menu";

let mockUserRole = "member";
const mockDeleteKeyResult = jest.fn();
const mockEditorRender = jest.fn();
const mockConfirmationRender = jest.fn();
const mockWriteText = jest.fn();

jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "key result" }),
  useUserRole: () => ({ userRole: mockUserRole }),
}));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn() },
}));
jest.mock("@/modules/objectives/hooks/use-delete-key-result-mutation", () => ({
  useDeleteKeyResultMutation: () => ({ mutate: mockDeleteKeyResult }),
}));
jest.mock("@/components/ui/confirm-dialog", () => ({
  ConfirmDialog: ({
    onConfirm,
    onClose,
  }: {
    onConfirm: () => void;
    onClose: () => void;
  }) => {
    mockConfirmationRender();
    return (
      <div aria-label="Delete key result" role="dialog">
        <button onClick={onConfirm} type="button">
          Confirm deletion
        </button>
        <button onClick={onClose} type="button">
          Cancel deletion
        </button>
      </div>
    );
  },
}));
jest.mock(
  "@/modules/objectives/stories/overview/update-key-result-dialog",
  () => ({
    UpdateKeyResultDialog: ({
      keyResult,
      onOpenChange,
      updateMode,
    }: {
      keyResult: KeyResult;
      onOpenChange: (open: boolean) => void;
      updateMode: string;
    }) => {
      mockEditorRender({ keyResult, updateMode });
      return (
        <div aria-label={`${updateMode} editor`} role="dialog">
          <button
            onClick={() => {
              onOpenChange(false);
            }}
            type="button"
          >
            Close editor
          </button>
        </div>
      );
    },
  }),
);

const KEY_RESULT: KeyResult = {
  id: "key result/1",
  sequenceId: 1,
  name: "Improve activation",
  measurementType: "percentage",
  objectiveId: "objective-1",
  startValue: 0,
  targetValue: 100,
  currentValue: 25,
  lead: null,
  contributors: [],
  startDate: "2026-10-01",
  endDate: "2026-12-31",
  createdAt: "2026-10-01T00:00:00Z",
  updatedAt: "2026-10-01T00:00:00Z",
  createdBy: "member-1",
};

// Deliberately opaque: the real map card does not forward injected DOM props.
const OpaqueCard = () => <button type="button">Improve activation</button>;

const renderMenu = (
  props: Partial<ComponentProps<typeof KeyResultContextMenu>> = {},
) => {
  const onOpenDetails = jest.fn();
  const result = render(
    <KeyResultContextMenu
      keyResult={KEY_RESULT}
      onOpenDetails={onOpenDetails}
      {...props}
    >
      <OpaqueCard />
    </KeyResultContextMenu>,
  );
  const card = screen.getByRole("button", { name: "Improve activation" });
  return { ...result, card, onOpenDetails };
};

const openContextMenu = (card: HTMLElement) => {
  fireEvent.contextMenu(card, { clientX: 70, clientY: 95 });
};

describe("key-result context menu", () => {
  beforeAll(() => {
    // jsdom does not provide the geometry constructor used by Radix's anchor.
    Object.defineProperty(globalThis, "DOMRect", {
      configurable: true,
      value: {
        fromRect: ({ x = 0, y = 0, width = 0, height = 0 }: DOMRectInit) => ({
          x,
          y,
          width,
          height,
          top: y,
          right: x + width,
          bottom: y + height,
          left: x,
        }),
      },
    });
  });

  beforeEach(() => {
    jest.clearAllMocks();
    mockUserRole = "member";
    mockWriteText.mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText: mockWriteText },
    });
  });

  it("opens the real Radix menu from an opaque card and invokes details once", () => {
    const { card, onOpenDetails } = renderMenu();
    expect(mockEditorRender).not.toHaveBeenCalled();
    expect(mockConfirmationRender).not.toHaveBeenCalled();

    openContextMenu(card);
    expect(screen.getByRole("menu")).toBeVisible();
    expect(onOpenDetails).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("menuitem", { name: "Open details" }));
    expect(onOpenDetails).toHaveBeenCalledTimes(1);
  });

  it.each([{ key: "ContextMenu" }, { key: "F10", shiftKey: true }])(
    "opens actions from a focused card using $key",
    (event) => {
      const { card } = renderMenu();
      card.focus();
      fireEvent.keyDown(card, event);
      expect(
        screen.getByRole("menuitem", { name: "Update progress..." }),
      ).toBeVisible();
    },
  );

  it.each([
    ["Edit...", "other"],
    ["Update progress...", "progress"],
  ])("mounts the correct editor only after selecting %s", (label, mode) => {
    const { card } = renderMenu();
    openContextMenu(card);
    fireEvent.click(screen.getByRole("menuitem", { name: label }));
    expect(
      screen.getByRole("dialog", { name: `${mode} editor` }),
    ).toBeVisible();
    expect(mockEditorRender).toHaveBeenLastCalledWith({
      keyResult: KEY_RESULT,
      updateMode: mode,
    });
    fireEvent.click(screen.getByRole("button", { name: "Close editor" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("requires confirmation before deleting and allows cancellation", () => {
    const { card } = renderMenu();
    openContextMenu(card);
    fireEvent.click(screen.getByRole("menuitem", { name: "Delete" }));
    expect(mockDeleteKeyResult).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Cancel deletion" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(mockDeleteKeyResult).not.toHaveBeenCalled();

    openContextMenu(card);
    fireEvent.click(screen.getByRole("menuitem", { name: "Delete" }));
    fireEvent.click(screen.getByRole("button", { name: "Confirm deletion" }));
    expect(mockDeleteKeyResult).toHaveBeenCalledWith({
      keyResultId: KEY_RESULT.id,
      objectiveId: KEY_RESULT.objectiveId,
    });
  });

  it("keeps read actions available while disabling every mutation for guests", () => {
    mockUserRole = "guest";
    const { card, onOpenDetails } = renderMenu();
    openContextMenu(card);
    for (const label of ["Edit...", "Update progress...", "Delete"]) {
      const item = screen.getByRole("menuitem", { name: label });
      expect(item).toHaveAttribute("aria-disabled", "true");
      fireEvent.click(item);
    }
    expect(mockEditorRender).not.toHaveBeenCalled();
    expect(mockConfirmationRender).not.toHaveBeenCalled();
    expect(mockDeleteKeyResult).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("menuitem", { name: "Open details" }));
    expect(onOpenDetails).toHaveBeenCalledTimes(1);
  });

  it("offers objective navigation and copies a workspace-aware key-result deep link", async () => {
    const objectivePath =
      "/acme/teams/team-1/objectives/objective-1?tab=stories";
    const { card } = renderMenu({ objectivePath });
    openContextMenu(card);
    expect(
      screen.getByRole("menuitem", { name: "Open objective" }),
    ).toHaveAttribute("href", objectivePath);
    fireEvent.click(screen.getByRole("menuitem", { name: "Copy link" }));
    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("Link copied to clipboard");
    });
    const copiedUrl = new URL(mockWriteText.mock.calls[0][0] as string);
    expect(copiedUrl.origin).toBe(window.location.origin);
    expect(copiedUrl.pathname).toBe(
      "/acme/teams/team-1/objectives/objective-1",
    );
    expect(copiedUrl.searchParams.get("tab")).toBe("overview");
    expect(copiedUrl.searchParams.get("keyResultId")).toBe(KEY_RESULT.id);
  });

  it("reports clipboard failures without claiming the link was copied", async () => {
    mockWriteText.mockRejectedValue(new Error("Clipboard denied"));
    const { card } = renderMenu({
      objectivePath: "/acme/teams/team-1/objectives/objective-1",
    });
    openContextMenu(card);
    fireEvent.click(screen.getByRole("menuitem", { name: "Copy link" }));
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        "Unable to copy link",
        expect.any(Object),
      );
    });
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("omits link actions when no objective route is supplied", () => {
    const { card } = renderMenu();
    openContextMenu(card);
    expect(
      screen.queryByRole("menuitem", { name: "Open objective" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("menuitem", { name: "Copy link" }),
    ).not.toBeInTheDocument();
  });
});
