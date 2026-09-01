import "server-only";

import type { UIMessage } from "ai";
import {
  isMutationCapableToolName,
  type MayaToolName,
} from "@/lib/ai/tool-policy";
import {
  ACTIVITY_PATTERN,
  ANALYTICS_PATH_TOOLS,
  ANALYTICS_PATTERN,
  ANALYTICS_SUPPORT_TOOLS,
  ATTACHMENT_PATTERN,
  BASE_TOOLS,
  CLARIFICATION_LANGUAGE_PATTERN,
  COMMENT_PATTERN,
  CONVERSATIONAL_REFERENCE_PATTERN,
  CREATE_PATTERN,
  DEFAULT_DISCOVERY_TOOLS,
  DELETE_PATTERN,
  DOCUMENT_PATTERN,
  FEEDBACK_PATTERN,
  FOCUS_PATTERN,
  FOLLOW_THROUGH_ACTION_PATTERN,
  GITHUB_PATTERN,
  GITHUB_READ_TOOLS,
  GITHUB_WRITE_TOOLS,
  INTEGRATION_REQUEST_PATTERN,
  INTEGRATION_REQUEST_READ_TOOLS,
  INTEGRATION_REQUEST_WRITE_TOOLS,
  LABEL_PATTERN,
  LINK_PATTERN,
  MEMORY_PATTERN,
  MUTATION_PATTERN,
  NAVIGATION_PATTERN,
  NOTIFICATION_PATTERN,
  OBJECTIVE_PATTERN,
  OBJECTIVE_READ_TOOLS,
  OBJECTIVE_WRITE_TOOLS,
  PLANNING_PATTERN,
  SEARCH_PATTERN,
  SPRINT_PATTERN,
  SPRINT_READ_TOOLS,
  STATUS_PATTERN,
  STORY_CREATE_PATTERN,
  STORY_CREATION_INTAKE_CLARIFICATION_PATTERN,
  STORY_CREATE_TOOLS,
  STORY_DELETE_TOOLS,
  STORY_PATTERN,
  STORY_PLANNING_CLARIFICATION_PATTERN,
  STORY_PLANNING_VALUE_PATTERN,
  STORY_READ_TOOLS,
  STORY_REFERENCE_PATTERN,
  STORY_UPDATE_TOOLS,
  TEAM_PATTERN,
  TEAM_READ_TOOLS,
  THEME_PATTERN,
  TOOL_DOMAIN_PROVENANCE,
  type ToolDomain,
  UPDATE_PATTERN,
} from "./policy-model";

const normalizeCustomStoryTerm = (term: string | undefined) => {
  if (!term) return [];

  const normalized = term
    .normalize("NFKC")
    .trim()
    .toLowerCase()
    .replace(/\s+/g, " ");
  if (
    normalized.length < 2 ||
    normalized.length > 48 ||
    !/^[\p{L}\p{N}]+(?:[ -][\p{L}\p{N}]+){0,3}$/u.test(normalized)
  ) {
    return [];
  }

  const terms = new Set([normalized]);
  if (normalized.endsWith("ies") && normalized.length > 3) {
    terms.add(`${normalized.slice(0, -3)}y`);
  } else if (normalized.endsWith("s") && normalized.length > 2) {
    terms.add(normalized.slice(0, -1));
  }

  return Array.from(terms);
};

const includesWholePhrase = (intent: string, phrase: string) => {
  const normalizedIntent = intent
    .normalize("NFKC")
    .toLowerCase()
    .replace(/[^\p{L}\p{N}-]+/gu, " ")
    .trim();
  return ` ${normalizedIntent} `.includes(` ${phrase} `);
};

const PATH_DOMAINS = [
  { pattern: /\/stories(?:\/|$)|\/my-work(?:\/|$)/, tools: STORY_READ_TOOLS },
  { pattern: /\/sprints(?:\/|$)/, tools: SPRINT_READ_TOOLS },
  { pattern: /\/objectives(?:\/|$)/, tools: OBJECTIVE_READ_TOOLS },
  { pattern: /\/requests(?:\/|$)/, tools: INTEGRATION_REQUEST_READ_TOOLS },
  { pattern: /\/feedback(?:\/|$)/, tools: ["listCustomerFeedbackTool"] },
  { pattern: /\/(?:docs|documents)(?:\/|$)/, tools: ["listDocumentsTool"] },
  { pattern: /\/analytics(?:\/|$)/, tools: ANALYTICS_PATH_TOOLS },
  { pattern: /\/notifications(?:\/|$)/, tools: ["notifications"] },
  { pattern: /\/teams(?:\/|$)/, tools: TEAM_READ_TOOLS },
] as const;

const CREATION_INTAKE_REDIRECTION_PATTERN =
  /\b(?:actually|instead|rather than|never mind|forget that|switch(?:ing)? to)\b/;
const NON_STORY_CREATION_CLARIFICATION_PATTERN =
  /\b(?:create|add)\s+(?:a|an|the|this|that)\s+(?:team|status|sprint|objective|key result|integration|connection|repository|memory|notification|document|label)\b|\b(?:team|status|sprint|objective|key result|integration|connection|repository|memory|notification|document|label)\b[^\n]{0,48}\b(?:called|named)\b/;
const CREATION_INTAKE_CANCELLATION_PATTERN = /\b(?:never mind|forget that)\b/;
const NON_STORY_CREATION_REQUEST_PATTERN =
  /\b(?:(?:create|add)\s+(?:(?:a|an|the|this|that)\s+)?|new\s+)(?:team|status|sprint|objective|key result|integration|connection|repository|memory|notification|document|label)\b/;
const EXPLICIT_DIFFERENT_DOMAIN_REQUEST_CUE_PATTERN =
  /\b(?:help me|can you|could you|configure|connect|set up|show me|list|find)\b/;
const EXPLICIT_DIFFERENT_DOMAIN_ACTION_PATTERN =
  /\b(?:make|start|draft|write|post|plan|schedule|open|read|get|check|review|reply|attach|upload|link|remember|notify)\b/;
const STORY_INTAKE_CORRECTION_CUE_PATTERN =
  /\b(?:use|call|set|change|assign|keep|skip|leave)\b/;
const STORY_INTAKE_REFERENCE_CORRECTION_ACTION_PATTERN =
  /\b(?:make|schedule|link)\b/;
const MAX_MARKED_SLOT_VALUE_LENGTH = 256;

const NON_STORY_TEXT_DOMAIN_PATTERNS = [
  { domain: "planning", pattern: PLANNING_PATTERN },
  { domain: "integration-request", pattern: INTEGRATION_REQUEST_PATTERN },
  { domain: "github", pattern: GITHUB_PATTERN },
  { domain: "objective", pattern: OBJECTIVE_PATTERN },
  { domain: "sprint", pattern: SPRINT_PATTERN },
  { domain: "status", pattern: STATUS_PATTERN },
  { domain: "team", pattern: TEAM_PATTERN },
  { domain: "memory", pattern: MEMORY_PATTERN },
  { domain: "notification", pattern: NOTIFICATION_PATTERN },
  { domain: "comment", pattern: COMMENT_PATTERN },
  { domain: "label", pattern: LABEL_PATTERN },
  { domain: "link", pattern: LINK_PATTERN },
  { domain: "attachment", pattern: ATTACHMENT_PATTERN },
  { domain: "feedback", pattern: FEEDBACK_PATTERN },
  { domain: "document", pattern: DOCUMENT_PATTERN },
] as const satisfies readonly {
  domain: Exclude<ToolDomain, "story">;
  pattern: RegExp;
}[];

const STORY_INTAKE_BOUNDARY_DOMAIN_PATTERNS = [
  ...NON_STORY_TEXT_DOMAIN_PATTERNS,
  {
    domain: "integration-request",
    pattern: /\b(?:integration|connection)\b/,
  },
] as const satisfies readonly {
  domain: Exclude<ToolDomain, "story">;
  pattern: RegExp;
}[];

const addTools = (
  selectedTools: Set<MayaToolName>,
  toolNames: readonly MayaToolName[],
) => {
  toolNames.forEach((toolName) => selectedTools.add(toolName));
};

const getMessageText = (message: UIMessage) =>
  message.parts
    .flatMap((part) => (part.type === "text" ? [part.text] : []))
    .join(" ")
    .toLowerCase();

const hasStoryCreationCorrectionIntent = (text: string) =>
  STORY_PLANNING_VALUE_PATTERN.test(text) ||
  ((CREATE_PATTERN.test(text) ||
    UPDATE_PATTERN.test(text) ||
    STORY_INTAKE_REFERENCE_CORRECTION_ACTION_PATTERN.test(text)) &&
    CONVERSATIONAL_REFERENCE_PATTERN.test(text)) ||
  (STORY_CREATION_INTAKE_CLARIFICATION_PATTERN.test(text) &&
    STORY_INTAKE_CORRECTION_CUE_PATTERN.test(text));

const isExplicitNonStoryDomainRequest = (text: string) => {
  const hasNonStoryDomain = STORY_INTAKE_BOUNDARY_DOMAIN_PATTERNS.some(
    ({ pattern }) => pattern.test(text),
  );
  if (!hasNonStoryDomain) return false;
  if (NON_STORY_CREATION_REQUEST_PATTERN.test(text)) return true;

  // Story intake uses several otherwise-shared domains as metadata. A direct
  // correction to its team, status, sprint, objective, label, or schedule must
  // not be mistaken for abandoning a story that does not exist yet.
  if (hasStoryCreationCorrectionIntent(text)) return false;

  return (
    MUTATION_PATTERN.test(text) ||
    EXPLICIT_DIFFERENT_DOMAIN_REQUEST_CUE_PATTERN.test(text) ||
    EXPLICIT_DIFFERENT_DOMAIN_ACTION_PATTERN.test(text)
  );
};

const isCreationIntakeRedirection = (text: string) => {
  if (isExplicitNonStoryDomainRequest(text)) return true;
  if (!CREATION_INTAKE_REDIRECTION_PATTERN.test(text)) return false;

  return (
    CREATION_INTAKE_CANCELLATION_PATTERN.test(text) &&
    !hasStoryCreationCorrectionIntent(text)
  );
};

const escapeRegExp = (value: string) =>
  value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

const hasMarkedSlotValue = (assistantText: string, slotValue: string) => {
  const normalizedSlotValue = slotValue.trim();
  if (
    normalizedSlotValue.length < 3 ||
    normalizedSlotValue.length > MAX_MARKED_SLOT_VALUE_LENGTH
  )
    return false;

  const escapedSlotValue = escapeRegExp(normalizedSlotValue);
  const wrappers = [
    ['"', '"'],
    ["“", "”"],
    ["'", "'"],
    ["**", "**"],
    ["__", "__"],
    ["`", "`"],
  ] as const;

  return wrappers.some(([opening, closing]) => {
    const markedValuePattern = new RegExp(
      `${escapeRegExp(opening)}\\s*${escapedSlotValue}\\s*${escapeRegExp(closing)}`,
      "i",
    );
    return markedValuePattern.test(assistantText);
  });
};

const getStoryCreationClarificationContext = ({
  acknowledgedSlotValue = "",
  assistantText,
  configuredStoryTerms,
}: {
  acknowledgedSlotValue?: string;
  assistantText: string;
  configuredStoryTerms: string[];
}) => {
  const assistantHasStoryContext =
    STORY_PATTERN.test(assistantText) ||
    configuredStoryTerms.some((term) =>
      includesWholePhrase(assistantText, term),
    );
  const assistantTextWithoutQuotedValues = assistantText.replace(
    /"[^"]*"|“[^”]*”|‘[^’]*’/g,
    " ",
  );
  const acknowledgesStorySlotValue = hasMarkedSlotValue(
    assistantText,
    acknowledgedSlotValue,
  );
  const hasConflictingCreationContext =
    !assistantHasStoryContext &&
    !acknowledgesStorySlotValue &&
    NON_STORY_CREATION_CLARIFICATION_PATTERN.test(
      assistantTextWithoutQuotedValues,
    );
  const isClarification = CLARIFICATION_LANGUAGE_PATTERN.test(assistantText);
  const asksForStoryMetadata =
    STORY_CREATION_INTAKE_CLARIFICATION_PATTERN.test(assistantText);
  const asksForPlanning =
    STORY_PLANNING_CLARIFICATION_PATTERN.test(assistantText);
  const assistantContinuesCreation =
    assistantHasStoryContext ||
    CREATE_PATTERN.test(assistantTextWithoutQuotedValues) ||
    CONVERSATIONAL_REFERENCE_PATTERN.test(assistantText) ||
    acknowledgesStorySlotValue;

  return {
    acknowledgesStorySlotValue,
    asksForPlanning,
    asksForStoryMetadata,
    assistantHasStoryContext,
    hasConflictingCreationContext,
    isClarification,
    isStoryCreationClarification:
      isClarification &&
      !hasConflictingCreationContext &&
      assistantContinuesCreation &&
      (asksForStoryMetadata || asksForPlanning),
  };
};

const isStoryCreationChainClarification = (
  clarification: ReturnType<typeof getStoryCreationClarificationContext>,
) =>
  clarification.isStoryCreationClarification ||
  (clarification.isClarification &&
    !clarification.hasConflictingCreationContext &&
    (clarification.asksForStoryMetadata || clarification.asksForPlanning));

const getLatestUserText = (messages: UIMessage[]) => {
  const message = messages.findLast((candidate) => candidate.role === "user");
  if (!message) return "";

  return getMessageText(message);
};

const getLatestAssistantContext = (messages: UIMessage[]) => {
  const latestUserMessageIndex = messages.findLastIndex(
    (message) => message.role === "user",
  );
  if (latestUserMessageIndex <= 0) {
    return { precedingUserText: "", text: "", toolNames: [] };
  }

  const latestAssistantMessage = messages[latestUserMessageIndex - 1];
  if (latestAssistantMessage.role !== "assistant") {
    return { precedingUserText: "", text: "", toolNames: [] };
  }

  const precedingUserMessage = messages
    .slice(0, latestUserMessageIndex - 1)
    .findLast((message) => message.role === "user");

  return {
    precedingUserText: precedingUserMessage
      ? getMessageText(precedingUserMessage)
      : "",
    text: getMessageText(latestAssistantMessage),
    toolNames: latestAssistantMessage.parts.flatMap((part) =>
      part.type.startsWith("tool-")
        ? [part.type.slice("tool-".length) as MayaToolName]
        : [],
    ),
  };
};

const isStoryCreationSlotValue = ({
  configuredStoryTerms,
  messageIndex,
  messages,
}: {
  configuredStoryTerms: string[];
  messageIndex: number;
  messages: UIMessage[];
}) => {
  if (messageIndex <= 0 || messageIndex >= messages.length - 1) return false;

  const slotMessage = messages[messageIndex];
  const precedingAssistantMessage = messages[messageIndex - 1];
  const followingAssistantMessage = messages[messageIndex + 1];
  if (
    slotMessage.role !== "user" ||
    precedingAssistantMessage.role !== "assistant" ||
    followingAssistantMessage.role !== "assistant"
  ) {
    return false;
  }

  const precedingClarification = getStoryCreationClarificationContext({
    assistantText: getMessageText(precedingAssistantMessage),
    configuredStoryTerms,
  });
  if (!isStoryCreationChainClarification(precedingClarification)) return false;

  const slotValue = getMessageText(slotMessage).trim();
  const followingAssistantText = getMessageText(followingAssistantMessage);
  const followingClarification = getStoryCreationClarificationContext({
    acknowledgedSlotValue: slotValue,
    assistantText: followingAssistantText,
    configuredStoryTerms,
  });
  const hasStrongStoryContinuation =
    followingClarification.assistantHasStoryContext ||
    followingClarification.asksForPlanning ||
    followingClarification.acknowledgesStorySlotValue;
  const hasBoundedMetadataContinuation =
    followingClarification.asksForStoryMetadata &&
    !isExplicitNonStoryDomainRequest(slotValue);

  return (
    followingClarification.isClarification &&
    !followingClarification.hasConflictingCreationContext &&
    (hasStrongStoryContinuation || hasBoundedMetadataContinuation)
  );
};

const hasRecentStoryCreationRequest = ({
  configuredStoryTerms,
  messages,
}: {
  configuredStoryTerms: string[];
  messages: UIMessage[];
}) => {
  const latestUserMessageIndex = messages.findLastIndex(
    (message) => message.role === "user",
  );

  // Chat context is bounded before tool selection. Scan that full recent chain
  // so the number of intake fields cannot make creation tools disappear.
  for (let index = latestUserMessageIndex - 1; index >= 0; index -= 1) {
    const message = messages[index];

    if (message.role === "assistant") {
      const hasCreationToolCall = message.parts.some(
        (part) =>
          part.type === "tool-createStory" ||
          part.type === "tool-bulkCreateStories",
      );
      if (hasCreationToolCall) return false;
      continue;
    }

    if (message.role !== "user") continue;

    const text = getMessageText(message);
    const hasConfiguredStoryCreationRequest = configuredStoryTerms.some(
      (term) => includesWholePhrase(text, term) && CREATE_PATTERN.test(text),
    );
    const isStoryCreationRequest =
      STORY_CREATE_PATTERN.test(text) || hasConfiguredStoryCreationRequest;

    if (isStoryCreationRequest) return true;
    if (
      isStoryCreationSlotValue({
        configuredStoryTerms,
        messageIndex: index,
        messages,
      })
    )
      continue;
    if (isCreationIntakeRedirection(text)) return false;

    // A creation anchor cannot cross an unverified user turn. This prevents a
    // generic clarification in a newer domain from reviving stale mutations.
    return false;
  }

  return false;
};

const isStoryCreationIntakeReply = ({
  acknowledgedSlotValue,
  assistantText,
  configuredStoryTerms,
  hasRecentCreationRequest,
  intent,
}: {
  acknowledgedSlotValue: string;
  assistantText: string;
  configuredStoryTerms: string[];
  hasRecentCreationRequest: boolean;
  intent: string;
}) => {
  if (
    !hasRecentCreationRequest ||
    !intent.trim() ||
    isCreationIntakeRedirection(intent) ||
    !CLARIFICATION_LANGUAGE_PATTERN.test(assistantText)
  ) {
    return false;
  }

  const clarification = getStoryCreationClarificationContext({
    acknowledgedSlotValue,
    assistantText,
    configuredStoryTerms,
  });
  const answersPlanningClarification =
    clarification.asksForPlanning &&
    (STORY_PLANNING_VALUE_PATTERN.test(intent) ||
      hasStoryCreationCorrectionIntent(intent));

  return (
    isStoryCreationChainClarification(clarification) &&
    (clarification.asksForStoryMetadata || answersPlanningClarification)
  );
};

const inferFollowUpDomain = ({
  assistantText,
  assistantToolNames,
  configuredStoryTerms,
  isFollowUp,
}: {
  assistantText: string;
  assistantToolNames: MayaToolName[];
  configuredStoryTerms: string[];
  isFollowUp: boolean;
}): ToolDomain | undefined => {
  if (!isFollowUp) return undefined;

  for (const toolName of assistantToolNames.toReversed()) {
    const matchedGroup = TOOL_DOMAIN_PROVENANCE.find(({ tools: toolNames }) =>
      toolNames.has(toolName),
    );
    if (matchedGroup) return matchedGroup.domain;
  }

  if (PLANNING_PATTERN.test(assistantText)) {
    return "planning";
  }

  if (
    configuredStoryTerms.some((term) =>
      includesWholePhrase(assistantText, term),
    )
  ) {
    return "story";
  }

  if (STORY_PATTERN.test(assistantText)) return "story";

  return NON_STORY_TEXT_DOMAIN_PATTERNS.find(({ pattern }) =>
    pattern.test(assistantText),
  )?.domain;
};

const getPendingMutationTools = (messages: UIMessage[]) => {
  const pendingTools = new Set<MayaToolName>();
  const latestAssistantMessage = messages.findLast(
    (message) => message.role === "assistant",
  );
  if (!latestAssistantMessage) return pendingTools;

  for (const part of latestAssistantMessage.parts) {
    if (!part.type.startsWith("tool-") || !("output" in part)) continue;

    const output = part.output;
    const needsConfirmation =
      output &&
      typeof output === "object" &&
      "needsConfirmation" in output &&
      output.needsConfirmation === true;
    const toolName = part.type.slice("tool-".length);

    if (needsConfirmation && isMutationCapableToolName(toolName)) {
      pendingTools.add(toolName as MayaToolName);
    }
  }

  return pendingTools;
};

const addAnalyticsTools = (
  selectedTools: Set<MayaToolName>,
  intent: string,
) => {
  addTools(selectedTools, ANALYTICS_SUPPORT_TOOLS);
  const isBroadWorkspaceRequest =
    /\bworkspace\b/.test(intent) &&
    /\b(?:analyze|analyse|analysis|health|insight|insights|overview|report|reports)\b/.test(
      intent,
    );

  if (/\bcommand center|dashboard\b/.test(intent)) {
    selectedTools.add("workspaceCommandCenterReportTool");
  } else if (isBroadWorkspaceRequest) {
    selectedTools.add("workspaceCommandCenterReportTool");
    selectedTools.add("workspacePerformanceReportTool");
  } else if (/\bpulse\b/.test(intent)) {
    selectedTools.add("pulseReportTool");
  } else if (/\bsprint|cycle|iteration\b/.test(intent)) {
    selectedTools.add("sprintPerformanceReportTool");
  } else if (/\bobjective|okr|key result|goal\b/.test(intent)) {
    selectedTools.add("objectiveProgressReportTool");
  } else if (/\bteam|member|person|people\b/.test(intent)) {
    selectedTools.add("teamPerformanceReportTool");
  } else if (
    /\b(?:story|stories|task|tasks|ticket|tickets|issue|issues|backlog)\b/.test(
      intent,
    )
  ) {
    selectedTools.add("storyPerformanceReportTool");
  } else if (/\btrend|trends|timeline\b/.test(intent)) {
    selectedTools.add("timelineTrendsReportTool");
  } else if (/\bworkload|capacity\b/.test(intent)) {
    selectedTools.add("workloadPlanningTool");
  } else if (/\bworkspace\b/.test(intent)) {
    selectedTools.add("workspaceCommandCenterReportTool");
    selectedTools.add("workspacePerformanceReportTool");
  } else {
    selectedTools.add("workspaceCommandCenterReportTool");
  }
};

export const selectActiveTools = ({
  currentPath = "",
  messages,
  storyTerminology,
}: {
  currentPath?: string;
  messages: UIMessage[];
  storyTerminology?: string;
}): MayaToolName[] => {
  const intent = getLatestUserText(messages);
  const configuredStoryTerms = normalizeCustomStoryTerm(storyTerminology);
  const hasConfiguredStoryTerm = configuredStoryTerms.some((term) =>
    includesWholePhrase(intent, term),
  );
  const hasStoryReference = STORY_REFERENCE_PATTERN.test(intent);
  const isStoryDetailPath = /\/story\/[^/?#]+/.test(currentPath.toLowerCase());
  const usesStoryDetailContext =
    isStoryDetailPath &&
    (ACTIVITY_PATTERN.test(intent) ||
      LABEL_PATTERN.test(intent) ||
      COMMENT_PATTERN.test(intent) ||
      LINK_PATTERN.test(intent) ||
      ATTACHMENT_PATTERN.test(intent) ||
      CONVERSATIONAL_REFERENCE_PATTERN.test(intent));
  const latestAssistantContext = getLatestAssistantContext(messages);
  const selectedTools = new Set<MayaToolName>(BASE_TOOLS);
  const pendingTools = getPendingMutationTools(messages);
  addTools(selectedTools, Array.from(pendingTools));

  // Legacy confirmation turns must keep the model constrained to the exact
  // prepared mutation. Native AI SDK approvals bypass this model-selection
  // path and are validated separately against persisted chat state.
  if (pendingTools.size > 0) return Array.from(selectedTools);

  let matchedDomain = false;
  const isFollowUp =
    CONVERSATIONAL_REFERENCE_PATTERN.test(intent) ||
    FOLLOW_THROUGH_ACTION_PATTERN.test(intent);
  const inferredDomain = inferFollowUpDomain({
    assistantText: latestAssistantContext.text,
    assistantToolNames: latestAssistantContext.toolNames,
    configuredStoryTerms,
    isFollowUp,
  });
  const isGenericFollowThrough = FOLLOW_THROUGH_ACTION_PATTERN.test(intent);
  const hasRecentCreationRequest = hasRecentStoryCreationRequest({
    configuredStoryTerms,
    messages,
  });
  const isStoryCreationIntake = isStoryCreationIntakeReply({
    acknowledgedSlotValue: latestAssistantContext.precedingUserText,
    assistantText: latestAssistantContext.text,
    configuredStoryTerms,
    hasRecentCreationRequest,
    intent,
  });
  const actionIntent = isGenericFollowThrough
    ? `${intent} ${latestAssistantContext.text}`
    : intent;
  const actionHasConfiguredStoryTerm = configuredStoryTerms.some((term) =>
    includesWholePhrase(actionIntent, term),
  );
  const isStoryIntent =
    STORY_PATTERN.test(intent) ||
    hasConfiguredStoryTerm ||
    inferredDomain === "story" ||
    isStoryCreationIntake ||
    hasStoryReference ||
    usesStoryDetailContext;
  const isStoryFollowThroughAction =
    pendingTools.size === 0 &&
    inferredDomain === "story" &&
    isGenericFollowThrough;
  const isStoryAction =
    isStoryIntent &&
    (STORY_CREATE_PATTERN.test(actionIntent) ||
      MUTATION_PATTERN.test(actionIntent) ||
      isStoryCreationIntake ||
      isStoryFollowThroughAction);

  if (isStoryIntent) {
    matchedDomain = true;
    addTools(selectedTools, STORY_READ_TOOLS);

    if (isStoryCreationIntake) {
      addTools(selectedTools, STORY_CREATE_TOOLS);
    } else {
      if (
        !DELETE_PATTERN.test(actionIntent) &&
        (STORY_CREATE_PATTERN.test(actionIntent) ||
          (actionHasConfiguredStoryTerm && CREATE_PATTERN.test(actionIntent)))
      )
        addTools(selectedTools, STORY_CREATE_TOOLS);
      if (UPDATE_PATTERN.test(actionIntent))
        addTools(selectedTools, STORY_UPDATE_TOOLS);
      if (DELETE_PATTERN.test(actionIntent))
        addTools(selectedTools, STORY_DELETE_TOOLS);
      if (/\bduplicate\b/.test(actionIntent))
        selectedTools.add("duplicateStory");
      if (/\brestore\b/.test(actionIntent)) selectedTools.add("restoreStory");
      if (/\bassociation|associate\b/.test(actionIntent)) {
        selectedTools.add("addStoryAssociation");
        selectedTools.add("removeStoryAssociation");
      }
      if (/\bcomment|reply\b/.test(actionIntent)) selectedTools.add("comments");
      if (/\blabel|tag\b/.test(actionIntent)) selectedTools.add("storyLabels");
      if (ACTIVITY_PATTERN.test(actionIntent))
        selectedTools.add("storyActivities");
      if (/\blink|url\b/.test(actionIntent)) selectedTools.add("links");
      if (/\battachment|file\b/.test(actionIntent)) {
        selectedTools.add("listAttachments");
        if (DELETE_PATTERN.test(actionIntent))
          selectedTools.add("deleteAttachment");
      }
    }
  }

  if (
    !isStoryAction &&
    (TEAM_PATTERN.test(intent) || inferredDomain === "team")
  ) {
    matchedDomain = true;
    addTools(selectedTools, TEAM_READ_TOOLS);
    const isTeamStatusIntent = STATUS_PATTERN.test(actionIntent);
    if (CREATE_PATTERN.test(actionIntent) && !isTeamStatusIntent) {
      selectedTools.add("createTeamTool");
      selectedTools.add("joinTeam");
    }
    if (UPDATE_PATTERN.test(actionIntent) && !isTeamStatusIntent)
      selectedTools.add("updateTeam");
    if (DELETE_PATTERN.test(actionIntent) && !isTeamStatusIntent) {
      selectedTools.add("deleteTeam");
      selectedTools.add("leaveTeam");
    }
  }

  if (
    !isStoryAction &&
    !OBJECTIVE_PATTERN.test(intent) &&
    (STATUS_PATTERN.test(intent) || inferredDomain === "status")
  ) {
    matchedDomain = true;
    selectedTools.add("statuses");
    selectedTools.add("listTeams");
  }

  if (
    !isStoryAction &&
    (SPRINT_PATTERN.test(intent) || inferredDomain === "sprint")
  ) {
    matchedDomain = true;
    addTools(selectedTools, SPRINT_READ_TOOLS);
    if (MUTATION_PATTERN.test(actionIntent))
      selectedTools.add("updateSprintSettings");
  }

  if (
    !isStoryAction &&
    (OBJECTIVE_PATTERN.test(intent) || inferredDomain === "objective")
  ) {
    matchedDomain = true;
    addTools(selectedTools, OBJECTIVE_READ_TOOLS);
    if (MUTATION_PATTERN.test(actionIntent))
      addTools(selectedTools, OBJECTIVE_WRITE_TOOLS);
  }

  if (FOCUS_PATTERN.test(intent)) {
    matchedDomain = true;
    selectedTools.add("focusBrief");
    selectedTools.add("resolveMember");
  } else if (ANALYTICS_PATTERN.test(intent)) {
    matchedDomain = true;
    addAnalyticsTools(selectedTools, intent);
  }

  if (
    !isStoryCreationIntake &&
    (PLANNING_PATTERN.test(intent) || inferredDomain === "planning")
  ) {
    matchedDomain = true;
    addTools(selectedTools, [
      "mayaWorkPlanTool",
      "applyMayaWorkPlanTool",
      "workloadPlanningTool",
      "focusBrief",
      "listTeamStories",
      "listTeamMembers",
      "resolveMember",
    ]);
  }

  if (GITHUB_PATTERN.test(intent) || inferredDomain === "github") {
    matchedDomain = true;
    addTools(selectedTools, GITHUB_READ_TOOLS);
    if (MUTATION_PATTERN.test(actionIntent))
      addTools(selectedTools, GITHUB_WRITE_TOOLS);
  }

  if (
    INTEGRATION_REQUEST_PATTERN.test(intent) ||
    inferredDomain === "integration-request"
  ) {
    matchedDomain = true;
    addTools(selectedTools, INTEGRATION_REQUEST_READ_TOOLS);
    if (MUTATION_PATTERN.test(actionIntent)) {
      addTools(selectedTools, INTEGRATION_REQUEST_WRITE_TOOLS);
    }
  }

  if (FEEDBACK_PATTERN.test(intent) || inferredDomain === "feedback") {
    matchedDomain = true;
    addTools(selectedTools, [
      "listCustomerFeedbackTool",
      "getCustomerFeedbackTool",
    ]);
  }

  if (DOCUMENT_PATTERN.test(intent) || inferredDomain === "document") {
    matchedDomain = true;
    addTools(selectedTools, ["listDocumentsTool", "getDocumentDetailsTool"]);
  }

  if (MEMORY_PATTERN.test(intent) || inferredDomain === "memory") {
    matchedDomain = true;
    selectedTools.add("listMemories");
    if (/\bremember|create|save|add\b/.test(actionIntent))
      selectedTools.add("createMemory");
    if (/\bupdate|edit|change\b/.test(actionIntent))
      selectedTools.add("updateMemory");
    if (/\bforget|delete|remove\b/.test(actionIntent))
      selectedTools.add("deleteMemory");
  }

  if (NOTIFICATION_PATTERN.test(intent) || inferredDomain === "notification") {
    matchedDomain = true;
    selectedTools.add("notifications");
  }

  if (
    ACTIVITY_PATTERN.test(intent) &&
    !hasStoryReference &&
    !usesStoryDetailContext
  ) {
    matchedDomain = true;
    selectedTools.add("activitySummaryTool");
  }

  if (
    !isStoryIntent &&
    (COMMENT_PATTERN.test(intent) || inferredDomain === "comment")
  ) {
    matchedDomain = true;
    selectedTools.add("comments");
  }

  if (
    !isStoryIntent &&
    (LABEL_PATTERN.test(intent) || inferredDomain === "label")
  ) {
    matchedDomain = true;
    selectedTools.add("labels");
  }

  if (
    !isStoryIntent &&
    (LINK_PATTERN.test(intent) || inferredDomain === "link")
  ) {
    matchedDomain = true;
    selectedTools.add("links");
  }

  if (
    !isStoryIntent &&
    (ATTACHMENT_PATTERN.test(intent) || inferredDomain === "attachment")
  ) {
    matchedDomain = true;
    selectedTools.add("listAttachments");
    if (DELETE_PATTERN.test(actionIntent))
      selectedTools.add("deleteAttachment");
  }

  if (NAVIGATION_PATTERN.test(intent)) selectedTools.add("navigation");
  if (THEME_PATTERN.test(intent)) selectedTools.add("theme");
  if (SEARCH_PATTERN.test(intent)) selectedTools.add("search");

  if (!matchedDomain) {
    const path = currentPath.toLowerCase();
    const pathDomain = PATH_DOMAINS.find(({ pattern }) => pattern.test(path));
    if (pathDomain) addTools(selectedTools, pathDomain.tools);
    else addTools(selectedTools, DEFAULT_DISCOVERY_TOOLS);
  }

  return Array.from(selectedTools);
};
