import type { ToolSet } from "ai";
import { compactToolOutput } from "./compact-tool-output";

export const withCompactModelOutputs = <TOOLS extends ToolSet>(
  toolSet: TOOLS,
): TOOLS =>
  Object.fromEntries(
    Object.entries(toolSet).map(([name, tool]) => [
      name,
      tool.toModelOutput
        ? tool
        : {
            ...tool,
            toModelOutput: ({ output }: { output: unknown }) => ({
              type: "json" as const,
              value: compactToolOutput(output),
            }),
          },
    ]),
  ) as TOOLS;
