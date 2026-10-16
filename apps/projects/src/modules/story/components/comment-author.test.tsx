import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import type { UserSummary } from "@/types/user-summary";
import { CommentAuthor } from "./comment-author";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));
jest.mock("@/lib/hooks/members", () => ({
  useMayaAssignee: () => ({ data: { id: "configured-maya" } }),
}));
jest.mock("ui", () => ({
  ...jest.requireActual("ui"),
  Tooltip: ({ title, children }: { title: ReactNode; children: ReactNode }) => (
    <>
      {title}
      {children}
    </>
  ),
}));

const user = (overrides: Partial<UserSummary> = {}): UserSummary => ({
  id: "configured-maya",
  username: "maya",
  fullName: "Maya",
  avatarUrl: null,
  isActive: true,
  isSystem: true,
  ...overrides,
});

it("identifies Maya's hover card by the configured actor ID after a display-name change", () => {
  const maya = user({ username: "assistant", fullName: "Renamed assistant" });
  render(<CommentAuthor user={maya} userId={maya.id} />);

  expect(screen.getByText("AI Agent")).toBeInTheDocument();
  expect(screen.queryByText("(System Account)")).not.toBeInTheDocument();
  expect(
    screen.queryByRole("link", { name: "Go to profile" }),
  ).not.toBeInTheDocument();
});

it("keeps other system accounts labeled System Account even when named Maya", () => {
  const system = user({ id: "other-system" });
  render(<CommentAuthor user={system} userId={system.id} />);

  expect(screen.getByText("(System Account)")).toBeInTheDocument();
  expect(screen.queryByText("AI Agent")).not.toBeInTheDocument();
});

it("keeps a human named Maya's profile action", () => {
  const human = user({ id: "human-maya", isSystem: false });
  render(<CommentAuthor user={human} userId={human.id} />);

  expect(screen.getByRole("link", { name: "Go to profile" })).toHaveAttribute(
    "href",
    "/acme/profile/human-maya",
  );
  expect(screen.queryByText("AI Agent")).not.toBeInTheDocument();
  expect(screen.queryByText("(System Account)")).not.toBeInTheDocument();
});
