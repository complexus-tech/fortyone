/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { ToolOutputRenderer } from "./tool-output-renderer";
import type { ToolMessagePart } from "./tool-output-policy";

jest.mock("@/lib/hooks/statuses", () => ({
  useStatuses: () => ({ data: [] }),
}));

jest.mock("../dot", () => ({ Dot: () => null }));
jest.mock("../priority-icon", () => ({ PriorityIcon: () => null }));
jest.mock("./entity-results", () => ({ EntityResults: () => null }));
jest.mock("./maya-work-plan-result", () => ({
  MayaWorkPlanResult: () => null,
}));
jest.mock("./story-results", () => ({ StoryResults: () => null }));

jest.mock("ui", () => ({
  Button: ({
    children,
    onClick,
  }: {
    children: ReactNode;
    onClick?: () => void;
  }) => (
    <button onClick={onClick} type="button">
      {children}
    </button>
  ),
  Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Tooltip: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

describe("ToolOutputRenderer", () => {
  it("renders a GitHub continuation control without exposing its signed URL", () => {
    const installUrl =
      "https://github.com/apps/fortyone/installations/new?state=signed-session";
    const part = {
      input: {},
      output: {
        installUrl,
        message: "GitHub install session created.",
        success: true,
      },
      state: "output-available",
      type: "tool-createGitHubInstallSessionTool",
    } as unknown as ToolMessagePart;
    const { container } = render(
      <ToolOutputRenderer
        onPromptSelect={jest.fn()}
        onToolApproval={jest.fn()}
        part={part}
      />,
    );

    expect(
      screen.getByRole("button", { name: "Continue to GitHub" }),
    ).toBeInTheDocument();
    expect(container).toHaveTextContent("GitHub install session created.");
    expect(container).not.toHaveTextContent("signed-session");
    expect(container.querySelector("a")).toBeNull();
  });
});
