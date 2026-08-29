export type RealtimeTheme = "dark" | "light" | "system" | "toggle";

export type RealtimeClientAction =
  | {
      path: string;
      type: "navigate";
    }
  | {
      theme: RealtimeTheme;
      type: "theme";
    };

const REALTIME_THEMES = new Set(["dark", "light", "system", "toggle"]);

export const extractRealtimeClientAction = <
  T extends {
    clientAction?: unknown;
  },
>(
  output: T,
): {
  action: RealtimeClientAction | null;
  modelOutput: Omit<T, "clientAction">;
} => {
  const { clientAction, ...modelOutput } = output;
  if (!clientAction || typeof clientAction !== "object") {
    return { action: null, modelOutput };
  }

  if (
    "type" in clientAction &&
    clientAction.type === "navigate" &&
    "path" in clientAction &&
    typeof clientAction.path === "string" &&
    clientAction.path.startsWith("/") &&
    !clientAction.path.startsWith("//")
  ) {
    return {
      action: { path: clientAction.path, type: "navigate" },
      modelOutput,
    };
  }

  if (
    "type" in clientAction &&
    clientAction.type === "theme" &&
    "theme" in clientAction &&
    typeof clientAction.theme === "string" &&
    REALTIME_THEMES.has(clientAction.theme)
  ) {
    return {
      action: {
        theme: clientAction.theme as RealtimeTheme,
        type: "theme",
      },
      modelOutput,
    };
  }

  return { action: null, modelOutput };
};
