export type VoiceTranscriptMessage = {
  id: string;
  role: "user" | "assistant";
  text: string;
  createdAt: number;
  order?: number;
};

export type VoicePendingAction = {
  id: string;
  title: string;
  description: string;
  details: readonly { label: string; value: string }[];
  isDestructive: boolean;
};

export type VoiceToolCall = {
  call_id: string;
  name: string;
  arguments: string;
};

export type VoiceToolResult = Record<string, unknown> & {
  success?: boolean;
  requiresConfirmation?: boolean;
  confirmationToken?: string;
  confirmation?: Record<string, unknown>;
  message?: string;
  error?: string;
};

export type VoiceEvent = {
  type?: string;
  item_id?: string;
  previous_item_id?: string | null;
  response_id?: string;
  delta?: string;
  transcript?: string;
  error?: { message?: string };
  item?: { id?: string; role?: string; type?: string };
  response?: {
    output?: {
      type?: string;
      call_id?: string;
      name?: string;
      arguments?: string;
    }[];
  };
};

const READ_TOOLS = new Set([
  "get_context",
  "list_teams",
  "list_team_members",
  "list_my_tasks",
  "search_work",
  "list_objectives",
  "list_key_results",
  "get_story",
  "sprints",
  "workload",
  "recent_activity",
  "customer_feedback",
  "workspace_briefing",
  "navigate",
  "set_theme",
  "end_conversation",
]);
const WRITE_TOOLS = new Set(["create_task", "update_story", "delete_story"]);

export function prepareVoiceTool(name: string, raw: string) {
  const parsed: unknown = raw.trim() ? JSON.parse(raw) : {};
  if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
    throw new Error("Maya proposed an invalid action. Please try again.");
  }
  const args = Object.fromEntries(
    Object.entries(parsed).filter(
      ([key]) =>
        !["confirmed", "confirmationtoken"].includes(key.toLowerCase()),
    ),
  );
  // A provider function call can never carry human approval, even if prompted
  // to invent one. The only confirmed request is made by approveAction().
  let mutation = WRITE_TOOLS.has(name);
  if (name === "story_comments") {
    const action = String(args.action ?? "list")
      .toLowerCase()
      .trim();
    if (action !== "list" && action !== "add")
      throw new Error("Unsupported comment action.");
    mutation = action === "add";
  } else if (name === "notifications") {
    const action = String(args.action ?? "list")
      .toLowerCase()
      .trim();
    if (
      !["list", "unread_count", "mark_read", "mark_all_read"].includes(action)
    ) {
      throw new Error("Unsupported inbox action.");
    }
    mutation = action === "mark_read" || action === "mark_all_read";
  } else if (!READ_TOOLS.has(name) && !mutation) {
    throw new Error("That action is not available in mobile voice.");
  }
  if (mutation) args.confirmed = false;
  return { args, mutation };
}

// Defense in depth: credentials are never copied into provider events, even if
// a future server response nests a confirmation payload more deeply.
export function redactVoiceToolResult(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(redactVoiceToolResult);
  if (!value || typeof value !== "object") return value;
  return Object.fromEntries(
    Object.entries(value)
      .filter(
        ([key]) =>
          ![
            "confirmationtoken",
            "clientsecret",
            "cookie",
            "authorization",
            "clientaction",
          ].includes(key.toLowerCase()),
      )
      .map(([key, entry]) => [key, redactVoiceToolResult(entry)]),
  );
}

export function voicePendingAction(
  id: string,
  name: string,
  args: Record<string, unknown>,
  output: VoiceToolResult,
): VoicePendingAction {
  const titles: Record<string, string> = {
    create_task: "Create task",
    update_story: "Update task",
    delete_story: "Delete task",
    story_comments: "Post comment",
    notifications: "Mark notifications read",
  };
  const details = Object.entries(output.confirmation ?? args)
    .filter(
      ([key, value]) =>
        !["confirmed", "confirmationToken"].includes(key) &&
        value !== undefined &&
        value !== null &&
        value !== "" &&
        value !== false,
    )
    .map(([key, value]) => ({
      label: key.replace(/([a-z])([A-Z])/g, "$1 $2").replace(/_/g, " "),
      value: typeof value === "string" ? value : JSON.stringify(value),
    }));
  return {
    id,
    title: titles[name] ?? "Confirm change",
    details,
    description:
      name === "delete_story"
        ? "Review this task before deleting it."
        : "Review the details before applying this change.",
    isDestructive: name === "delete_story",
  };
}

export function voiceTranscriptUpdate(event: VoiceEvent) {
  const type = event.type ?? "";
  const user = type.startsWith("conversation.item.input_audio_transcription.");
  const assistant =
    /^(response\.(output_audio_transcript|audio_transcript))\.(delta|done)$/.test(
      type,
    );
  if (!user && !assistant) return null;
  const id = event.item_id ?? event.response_id;
  if (!id) return null;
  const final = type.endsWith(".completed") || type.endsWith(".done");
  const text = final ? event.transcript : event.delta;
  if (typeof text !== "string") return null;
  return {
    id,
    text,
    final,
    role: user ? ("user" as const) : ("assistant" as const),
  };
}

export function voiceConversationContext(
  messages: readonly (Pick<VoiceTranscriptMessage, "role" | "text"> & {
    id?: string;
  })[],
) {
  const seen = new Set<string>();
  return messages
    .filter((message) => {
      if (!message.text.trim() || (message.id && seen.has(message.id)))
        return false;
      if (message.id) seen.add(message.id);
      return true;
    })
    .slice(-24)
    .map(({ role, text }) => ({
      role,
      text: Array.from(text.trim()).slice(0, 4000).join(""),
    }));
}

/** The deployed voice endpoint accepts the same two fields as the web client. */
export function voiceSessionRequest(context: {
  currentPath: string;
  messages: readonly Pick<VoiceTranscriptMessage, "role" | "text">[];
}) {
  return {
    currentPath: context.currentPath,
    messages: context.messages.map(({ role, text }) => ({ role, text })),
  };
}

const UUID =
  "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}";
export function nativeVoiceRoute(path: string): string | null {
  const staticPaths: Record<string, string> = {
    "/my-work": "/my-work",
    "/notifications": "/inbox",
    "/summary": "/",
    "/settings": "/settings",
  };
  if (Object.hasOwn(staticPaths, path)) return staticPaths[path];
  const story = path.match(new RegExp(`^/work/(${UUID})$`));
  if (story) return `/story/${story[1]}`;
  const team = path.match(
    new RegExp(`^/teams/(${UUID})/(stories|objectives|sprints)$`),
  );
  if (team)
    return `/teams/${team[1]}${team[2] === "stories" ? "" : `/${team[2]}`}`;
  const objective = path.match(
    new RegExp(`^/teams/(${UUID})/objectives/(${UUID})$`),
  );
  if (objective) return `/team/${objective[1]}/objectives/${objective[2]}`;
  const sprint = path.match(
    new RegExp(`^/teams/(${UUID})/sprints/(${UUID})/stories$`),
  );
  if (sprint) return `/team/${sprint[1]}/sprints/${sprint[2]}`;
  return null;
}
