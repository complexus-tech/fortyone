import type { MayaMessage } from "../types";
import { asRecord, isEntityId } from "./chat-protocol";

export type MayaClientAction =
  | {
      type: "navigate";
      href: "/(tabs)" | "/(tabs)/my-work" | "/(tabs)/inbox" | "/settings";
    }
  | {
      type: "navigate";
      href: { pathname: "/story/[storyId]"; params: { storyId: string } };
    }
  | {
      type: "navigate";
      href: {
        pathname: "/teams/[teamId]";
        params: { teamId: string; sprintId?: string; objectiveId?: string };
      };
    }
  | { type: "theme"; theme: "light" | "dark" | "system" | "toggle" };

/** Use structured tool inputs, never follow a model-supplied URL. */
export const getMayaClientActions = (
  message: MayaMessage,
): MayaClientAction[] =>
  message.parts.flatMap((part): MayaClientAction[] => {
    const tool = asRecord(part);
    const input = asRecord(tool.input);
    const output = asRecord(tool.output);
    if (
      tool.state !== "output-available" ||
      output.error ||
      output.success === false
    )
      return [];
    if (
      part.type === "tool-theme" &&
      ["light", "dark", "system", "toggle"].includes(String(output.theme))
    ) {
      return [
        {
          type: "theme",
          theme: output.theme as "light" | "dark" | "system" | "toggle",
        },
      ];
    }
    if (part.type !== "tool-navigation") return [];
    switch (input.targetType) {
      case "summary":
        return [{ type: "navigate", href: "/(tabs)" }];
      case "my-work":
        return [{ type: "navigate", href: "/(tabs)/my-work" }];
      case "notifications":
        return [{ type: "navigate", href: "/(tabs)/inbox" }];
      case "settings":
        return [{ type: "navigate", href: "/settings" }];
      case "story":
        return isEntityId(input.entityId)
          ? [
              {
                type: "navigate",
                href: {
                  pathname: "/story/[storyId]",
                  params: { storyId: input.entityId },
                },
              },
            ]
          : [];
      case "team":
        return isEntityId(input.teamId) &&
          (!input.route || input.route === "stories")
          ? [
              {
                type: "navigate",
                href: {
                  pathname: "/teams/[teamId]",
                  params: { teamId: input.teamId },
                },
              },
            ]
          : [];
      case "sprint":
      case "objective":
        return isEntityId(input.teamId) && isEntityId(input.entityId)
          ? [
              {
                type: "navigate",
                href: {
                  pathname: "/teams/[teamId]",
                  params: {
                    teamId: input.teamId,
                    [input.targetType === "sprint"
                      ? "sprintId"
                      : "objectiveId"]: input.entityId,
                  },
                },
              },
            ]
          : [];
      default:
        return [];
    }
  });

export const hasMayaMutationReceipt = (message: MayaMessage) =>
  message.parts.some((part) => {
    const tool = asRecord(part);
    return (
      part.type.startsWith("tool-") &&
      (tool.state === "output-available" || tool.state === "output-error") &&
      asRecord(tool.approval).approved === true
    );
  });
