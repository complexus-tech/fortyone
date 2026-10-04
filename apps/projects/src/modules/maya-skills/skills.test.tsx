import type { ComponentProps } from "react";
import { useState } from "react";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { MayaSkillEditor } from "./skill-editor";
import { MayaSkillPicker } from "./skill-picker";
import { MayaSkillsDialog } from "./skills-dialog";
import { insertMayaSkillInstructions } from "./starters";
import type { MayaSkill } from "./types";

const mockCreate = jest.fn();
const mockUpdate = jest.fn();
const mockDelete = jest.fn();
const mockQuery = jest.fn();
const mockRefetch = jest.fn();
const mockReset = jest.fn();
let mockPending = false;
let mockWorkspace = "acme";
let mockUser = "user-one";
let mockDeleteError: Error | null = null;

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUser } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspace }),
}));
jest.mock("./hooks", () => ({
  useMayaSkills: (...args: unknown[]) => mockQuery(...args),
  useMayaSkillMutations: () => ({
    create: { mutateAsync: mockCreate, isPending: mockPending },
    update: { mutateAsync: mockUpdate, isPending: mockPending },
    remove: {
      mutateAsync: mockDelete,
      isPending: mockPending,
      error: mockDeleteError,
      reset: mockReset,
    },
  }),
}));
jest.mock("sonner", () => ({ toast: { success: jest.fn() } }));

const skill: MayaSkill = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  name: "Weekly update",
  description: "Summarize team progress",
  instructions: "Review current work and draft an update.",
  createdAt: "2026-10-04T09:00:00Z",
  updatedAt: "2026-10-04T09:00:00Z",
};
const loaded = {
  data: [skill],
  isPending: false,
  isError: false,
  isFetching: false,
  refetch: mockRefetch,
};
const resizeDescriptor = Object.getOwnPropertyDescriptor(
  globalThis,
  "ResizeObserver",
);
const scrollDescriptor = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  "scrollIntoView",
);
beforeAll(() => {
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: class {
      observe() {}
      unobserve() {}
      disconnect() {}
    },
  });
  Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
    configurable: true,
    value: jest.fn(),
  });
});
afterAll(() => {
  if (resizeDescriptor)
    Object.defineProperty(globalThis, "ResizeObserver", resizeDescriptor);
  else Reflect.deleteProperty(globalThis, "ResizeObserver");
  if (scrollDescriptor)
    Object.defineProperty(
      HTMLElement.prototype,
      "scrollIntoView",
      scrollDescriptor,
    );
  else Reflect.deleteProperty(HTMLElement.prototype, "scrollIntoView");
});
beforeEach(() => {
  mockCreate.mockReset().mockResolvedValue(skill);
  mockUpdate.mockReset().mockResolvedValue(skill);
  mockDelete.mockReset().mockResolvedValue(null);
  mockRefetch.mockReset().mockResolvedValue({});
  mockQuery.mockReset().mockReturnValue(loaded);
  mockPending = false;
  mockWorkspace = "acme";
  mockUser = "user-one";
  mockDeleteError = null;
});

const Picker = (props: Partial<ComponentProps<typeof MayaSkillPicker>>) => {
  const [open, setOpen] = useState(false);
  return (
    <MayaSkillPicker
      disabled={false}
      onOpenChange={setOpen}
      onReturnFocus={jest.fn()}
      onValueChange={jest.fn()}
      open={open}
      value=""
      {...props}
    />
  );
};

it("searches and selects by keyboard without mutating a saved skill", async () => {
  const select = jest.fn();
  render(<Picker onValueChange={select} />);
  fireEvent.click(screen.getByRole("button", { name: "Skills" }));
  const search = await screen.findByLabelText("Search skills");
  expect(search).toHaveFocus();
  fireEvent.change(search, { target: { value: "team progress" } });
  expect(select).not.toHaveBeenCalled();
  fireEvent.keyDown(search, { key: "Enter" });
  expect(select).toHaveBeenCalledWith(skill.instructions);
  expect(mockCreate).not.toHaveBeenCalled();
  expect(mockUpdate).not.toHaveBeenCalled();
});

it("preserves an existing draft when inserting a skill and replaces only an empty slash entry", () => {
  expect(
    insertMayaSkillInstructions("For the support team", skill.instructions),
  ).toBe(`For the support team\n\n${skill.instructions}`);
  expect(insertMayaSkillInstructions("/", skill.instructions)).toBe(
    skill.instructions,
  );
  expect(
    insertMayaSkillInstructions(
      "/a meaningful existing prompt",
      skill.instructions,
    ),
  ).toContain("/a meaningful existing prompt\n\n");
});

it("keeps starters unsaved until the user reviews and creates one", async () => {
  render(<MayaSkillsDialog onClose={jest.fn()} />);
  expect(mockCreate).not.toHaveBeenCalled();
  fireEvent.click(
    screen.getByRole("button", { name: "Use example: Plan my week" }),
  );
  const editor = screen.getByRole("dialog", { name: "Create skill" });
  expect(within(editor).getByLabelText("Skill name")).toHaveValue(
    "Plan my week",
  );
  expect(mockCreate).not.toHaveBeenCalled();
  fireEvent.click(within(editor).getByRole("button", { name: "Create skill" }));
  await waitFor(() => {
    expect(mockCreate).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "Plan my week",
        instructions: expect.stringContaining("available capacity"),
      }),
    );
  });
});

it("retains a failed edit and sends the original concurrency version", async () => {
  mockUpdate.mockRejectedValue(
    new Error("The skill changed elsewhere. Reopen it before saving."),
  );
  const close = jest.fn();
  render(<MayaSkillEditor onClose={close} skill={skill} />);
  fireEvent.change(screen.getByLabelText("Instructions"), {
    target: { value: "My updated instructions" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Save skill" }));
  await waitFor(() => {
    expect(screen.getByRole("alert")).toHaveTextContent("changed elsewhere");
  });
  expect(screen.getByLabelText("Instructions")).toHaveValue(
    "My updated instructions",
  );
  expect(mockUpdate).toHaveBeenCalledWith({
    id: skill.id,
    input: {
      name: skill.name,
      description: skill.description,
      instructions: "My updated instructions",
      updatedAt: skill.updatedAt,
    },
  });
  expect(close).not.toHaveBeenCalled();
});

it("keeps overlong Unicode drafts intact and explains the character limit", async () => {
  render(<MayaSkillEditor onClose={jest.fn()} skill={null} />);
  const name = screen.getByLabelText("Skill name");
  const description = screen.getByLabelText("Description (optional)");
  const instructions = screen.getByLabelText("Instructions");
  for (const control of [name, description, instructions]) {
    expect(control).not.toHaveAttribute("maxLength");
  }
  const overlongName = "😀".repeat(81);
  fireEvent.change(name, { target: { value: overlongName } });
  fireEvent.change(instructions, { target: { value: "Review my tasks." } });
  fireEvent.click(screen.getByRole("button", { name: "Create skill" }));
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Keep the name to 80 characters or fewer.",
  );
  expect(name).toHaveValue(overlongName);
  expect(mockCreate).not.toHaveBeenCalled();
  const validName = "😀".repeat(80);
  fireEvent.change(name, { target: { value: validName } });
  fireEvent.click(screen.getByRole("button", { name: "Create skill" }));
  await waitFor(() => {
    expect(mockCreate).toHaveBeenCalledWith({
      name: validName,
      description: "",
      instructions: "Review my tasks.",
    });
  });
});

it("keeps a conflicted draft intact and reopens the refreshed skill version", async () => {
  mockUpdate.mockRejectedValue(
    new Error("Skill changed elsewhere. Reopen it before saving."),
  );
  const { rerender } = render(<MayaSkillsDialog onClose={jest.fn()} />);
  const openEditor = async () => {
    fireEvent.pointerDown(
      screen.getByRole("button", { name: "Actions for Weekly update" }),
      { button: 0, ctrlKey: false },
    );
    fireEvent.click(
      await screen.findByRole("menuitem", { name: "Edit skill" }),
    );
  };
  await openEditor();
  fireEvent.change(screen.getByLabelText("Instructions"), {
    target: { value: "My unsaved draft" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Save skill" }));
  await screen.findByRole("alert");
  const latest = {
    ...skill,
    instructions: "Instructions saved in another tab",
    updatedAt: "2026-10-04T10:00:00Z",
  };
  mockQuery.mockReturnValue({ ...loaded, data: [latest] });
  rerender(<MayaSkillsDialog onClose={jest.fn()} />);
  expect(screen.getByLabelText("Instructions")).toHaveValue("My unsaved draft");
  fireEvent.click(
    within(screen.getByRole("dialog", { name: "Edit skill" })).getByRole(
      "button",
      { name: "Cancel" },
    ),
  );
  await openEditor();
  expect(screen.getByLabelText("Instructions")).toHaveValue(
    latest.instructions,
  );
  mockUpdate.mockResolvedValue(latest);
  fireEvent.click(screen.getByRole("button", { name: "Save skill" }));
  await waitFor(() => {
    expect(mockUpdate).toHaveBeenLastCalledWith({
      id: latest.id,
      input: {
        name: latest.name,
        description: latest.description,
        instructions: latest.instructions,
        updatedAt: latest.updatedAt,
      },
    });
  });
});

it("prevents dismissal and duplicate saves while a mutation is pending", () => {
  mockPending = true;
  const close = jest.fn();
  render(<MayaSkillEditor onClose={close} skill={skill} />);
  const dialog = screen.getByRole("dialog", { name: "Edit skill" });
  expect(dialog).toHaveAttribute("aria-busy", "true");
  expect(screen.getByLabelText("Instructions")).toBeDisabled();
  expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
  fireEvent.keyDown(dialog, { key: "Escape" });
  fireEvent.pointerDown(document.body, { button: 0, pointerType: "mouse" });
  fireEvent.click(screen.getByRole("button", { name: "Saving..." }));
  expect(close).not.toHaveBeenCalled();
  expect(mockUpdate).not.toHaveBeenCalled();
});

it("shows fetch failures with a retry and prevents selection when disabled", async () => {
  mockQuery.mockReturnValue({ ...loaded, isError: true });
  const select = jest.fn();
  const { rerender } = render(<Picker onValueChange={select} />);
  fireEvent.click(screen.getByRole("button", { name: "Skills" }));
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "could not be loaded",
  );
  fireEvent.click(screen.getByRole("button", { name: "Try again" }));
  expect(mockRefetch).toHaveBeenCalledTimes(1);
  rerender(<Picker disabled onValueChange={select} />);
  expect(screen.queryByLabelText("Search skills")).not.toBeInTheDocument();
  expect(screen.getByRole("button", { name: "Skills" })).toBeDisabled();
  expect(select).not.toHaveBeenCalled();
});

it("closes personal management when the authenticated workspace changes", async () => {
  const { rerender } = render(<Picker />);
  fireEvent.click(screen.getByRole("button", { name: "Skills" }));
  fireEvent.click(await screen.findByRole("button", { name: "Manage skills" }));
  expect(screen.getByRole("dialog", { name: "Maya skills" })).toBeVisible();
  mockWorkspace = "other";
  rerender(<Picker />);
  await waitFor(() => {
    expect(
      screen.queryByRole("dialog", { name: "Maya skills" }),
    ).not.toBeInTheDocument();
  });
});

it("deletes only after explicit confirmation and retains errors for retry", async () => {
  mockDelete.mockRejectedValue(new Error("Connection lost"));
  const { rerender } = render(<MayaSkillsDialog onClose={jest.fn()} />);
  fireEvent.pointerDown(
    screen.getByRole("button", { name: "Actions for Weekly update" }),
    { button: 0, ctrlKey: false },
  );
  fireEvent.click(
    await screen.findByRole("menuitem", { name: "Delete skill" }),
  );
  expect(mockDelete).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Delete skill" }));
  await waitFor(() => {
    expect(mockDelete).toHaveBeenCalledWith(skill.id);
  });
  mockDeleteError = new Error("Connection lost");
  await act(async () => {
    rerender(<MayaSkillsDialog onClose={jest.fn()} />);
  });
  expect(screen.getByRole("alert")).toHaveTextContent("Connection lost");
  expect(screen.getByRole("dialog", { name: "Delete skill" })).toBeVisible();
});
