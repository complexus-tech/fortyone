import type { UIMessage } from "ai";
import { MAYA_TOOL_NAMES } from "@/lib/ai/tool-names";
import { selectActiveToolPlan, selectActiveTools } from "./active-tools";

const userMessage = (text: string): UIMessage => ({
  id: `user-${text}`,
  role: "user",
  parts: [{ type: "text", text }],
});

describe("Maya active tools", () => {
  it.each([
    "Create a story and assign it to me",
    "Gadzira story itsva ugondipa",
    "Crea una historia y asígnamela",
    "أنشئ قصة وأسندها إليّ",
    "ツールを確認して、このストーリーを編集して",
  ])("exposes the complete registry independently of language: %s", (text) => {
    expect(
      selectActiveTools({
        currentPath: "/acme/stories",
        messages: [userMessage(text)],
        storyTerminology: "work item",
      }),
    ).toEqual(MAYA_TOOL_NAMES);
  });

  it("does not let the previous domain suppress a new topic", () => {
    const messages: UIMessage[] = [
      userMessage("Show me the objective Improve Customer Value"),
      {
        id: "assistant-objective",
        role: "assistant",
        parts: [{ type: "text", text: "Here is the objective." }],
      },
      userMessage("Now switch the theme to dark"),
      {
        id: "assistant-theme",
        role: "assistant",
        parts: [{ type: "text", text: "Done." }],
      },
      userMessage("Bvisa story yandanga ndichitaura nezvayo"),
    ];

    const plan = selectActiveToolPlan({
      currentPath: "/acme/objectives",
      messages,
    });

    expect(plan.source).toBe("universal");
    expect(plan.activeTools).toEqual(MAYA_TOOL_NAMES);
    expect(plan.activeTools).toEqual(
      expect.arrayContaining([
        "getObjectiveDetailsTool",
        "theme",
        "deleteStory",
      ]),
    );
  });

  it("returns a fresh list without mutating the canonical registry", () => {
    const first = selectActiveTools({ messages: [userMessage("hello")] });
    first.pop();

    expect(selectActiveTools({ messages: [userMessage("hello again")] })).toEqual(
      MAYA_TOOL_NAMES,
    );
  });
});
