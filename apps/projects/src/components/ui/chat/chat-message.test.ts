/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("ChatMessage", () => {
  it("identifies Maya as FortyOne's AI agent", () => {
    const promptSource = readSource("src/app/api/chat/system.ts");

    expect(promptSource).toContain(
      "You are Maya, FortyOne's AI agent for project management.",
    );
    expect(promptSource.length).toBeLessThan(10_000);
  });

  it("uses story-list surface colors for user prompts instead of inverse colors", () => {
    const source = readSource("src/components/ui/chat/chat-message.tsx");

    expect(source).toContain("bg-state-hover/80");
    expect(source).toContain("dark:bg-white/[0.08]");
    expect(source).not.toContain("text-foreground-inverse");
    expect(source).not.toContain("bg-background-inverse rounded-tr-md");
  });

  it("lets short user prompts use their intrinsic width", () => {
    const source = readSource("src/components/ui/chat/chat-message.tsx");

    expect(source).toContain("w-fit max-w-[80%] flex-none items-end");
    expect(source).toContain(
      "bg-state-hover/80 w-fit rounded-tr-md break-words",
    );
    expect(source).toContain("max-w-full flex-1");
  });

  it("clamps long user prompts to 20 lines with an accessible toggle", () => {
    const source = readSource("src/components/ui/chat/chat-message.tsx");

    expect(source).toContain("const USER_PROMPT_MAX_LINES = 20");
    expect(source).toContain("aria-expanded={isExpanded}");
    expect(source).toContain('"Show less" : "Show more"');
    expect(source).toContain('"overflow-hidden": !isExpanded');
  });

  it("renders detected user-prompt links as safe external links", () => {
    const source = readSource("src/components/ui/chat/chat-message.tsx");

    expect(source).toContain("getPromptTextSegments(text)");
    expect(source).toContain('rel="noopener noreferrer"');
    expect(source).toContain('target="_blank"');
    expect(source).toContain("underline underline-offset-2");
  });

  it("does not force text colors in chat message markdown", () => {
    const source = readSource("src/styles/global.css");

    expect(source).not.toMatch(/\.chat-tables[\s\S]*text-foreground/);
  });

  it("lets the chat composer inherit theme text color", () => {
    const source = readSource("src/components/ui/chat/chat-input.tsx");

    expect(source).not.toContain("dark:text-white");
  });

  it("renders assistant links as plain text", () => {
    const messageSource = readSource("src/components/ui/chat/chat-message.tsx");
    const promptSource = readSource("src/app/api/chat/system.ts");

    expect(messageSource).toContain("a: LinkText");
    expect(messageSource).toContain("components={STREAMDOWN_COMPONENTS}");
    expect(promptSource).toContain(
      "Do not embed internal FortyOne links in responses",
    );
    expect(promptSource).not.toContain(
      "link its human-readable reference or title",
    );
  });

  it("renders text and supported tool outputs in their original part order", () => {
    const source = readSource("src/components/ui/chat/chat-message.tsx");
    const utilsSource = readSource(
      "src/components/ui/chat/chat-message-utils.ts",
    );

    expect(source).toContain("message.parts.map");
    expect(source).toContain("<ToolOutputRenderer");
    expect(utilsSource).toContain("isRenderableToolPart(part)");
    expect(source).not.toContain("const reportParts");
    expect(source).not.toContain("const suggestionParts");
  });

  it("renders one polished generative report for a specific sprint", () => {
    const rendererSource = readSource(
      "src/components/ui/chat/tool-output-renderer.tsx",
    );
    const sprintReportSource = readSource(
      "src/components/ui/chat/analytics-report/sprint-report.tsx",
    );
    const promptSource = readSource("src/app/api/chat/system.ts");

    expect(sprintReportSource).toContain("<BurndownChart");
    expect(sprintReportSource).toContain("workingDays={asWorkingDays");
    expect(sprintReportSource).toContain('title="Team allocation"');
    expect(sprintReportSource).toContain(
      "Tracks remaining work against the ideal sprint pace.",
    );
    expect(sprintReportSource).toContain(
      "Shows completed and assigned work for each team member.",
    );
    expect(sprintReportSource).toContain(".slice(0, 5)");
    expect(sprintReportSource).toContain("maxLabelLength={12}");
    expect(rendererSource).not.toContain("getSprintBurndownData");
    expect(rendererSource).not.toContain("Burndown graph");
    expect(promptSource).toContain(
      "After a single-sprint generative report, add at most one brief interpretation",
    );
  });

  it("tells Maya never to duplicate generative UI data in text", () => {
    const promptSource = readSource("src/app/api/chat/system.ts");

    expect(promptSource).toContain(
      "Treat every generative UI result as the canonical, complete presentation of its data",
    );
    expect(promptSource).toContain(
      "Never repeat, enumerate, summarize, or reformat information already visible in generative UI",
    );
    expect(promptSource).toContain(
      "normally return no follow-up text after the UI",
    );
    expect(promptSource).toContain(
      "User-facing generative UI tools are presentation tools",
    );
    expect(promptSource).toContain(
      "When stories are only evidence for a comparison, duplicate check, classification, review, or recommendation, use the search tool",
    );
    expect(promptSource).toContain(
      "Empty interactive-list results appear as one plain no-results sentence instead of generative UI",
    );
  });

  it("uses private evidence for focus advice and reserves reports for explicit intent", () => {
    const promptSource = readSource("src/app/api/chat/system.ts");

    expect(promptSource).toContain(
      "Default to at most one user-facing generative UI result per response",
    );
    expect(promptSource).toContain(
      'For advice such as "what should I focus on today/next?", "what needs attention?", or "what should this person/team focus on?", use focusBrief',
    );
    expect(promptSource).toContain(
      "Treat focusBrief as private evidence. Never expose or describe its payload",
    );
    expect(promptSource).toContain(
      "Do not infer a visual report from a request for advice",
    );
    expect(promptSource).toContain(
      "Use resolveMember when a member name or username must be converted to an ID",
    );
  });

  it("limits my-team context and clarification to joined teams", () => {
    const chatHookSource = readSource(
      "src/modules/maya/hooks/use-maya-chat.ts",
    );
    const routeSource = readSource("src/app/api/chat/route.ts");
    const promptSource = readSource("src/app/api/chat/system.ts");

    expect(chatHookSource).not.toContain("useTeams()");
    expect(chatHookSource).not.toContain("useJoinedTeams()");
    expect(routeSource).toContain("resolveJoinedTeams");
    expect(routeSource).not.toContain("joinedTeams = []");
    expect(promptSource).toContain(
      '"My team" and "our team" always mean teams the user has joined',
    );
    expect(promptSource).toContain(
      'Never offer a public-but-unjoined team as a clarification option for "my team" or "our team"',
    );
  });

  it("renders empty list results as plain text without generative list chrome", () => {
    const source = readSource("src/components/ui/chat/generative-list.tsx");

    expect(source).toContain("if (items.length === 0)");
    expect(source).toContain(
      '<p className="text-text-muted text-base">{emptyMessage}</p>',
    );
    expect(source).not.toContain("{items.length ?");
  });
});
