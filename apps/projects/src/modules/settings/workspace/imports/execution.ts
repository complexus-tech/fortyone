import type { Member } from "@/types/member";
import type { State, StateCategory } from "@/types/states";
import {
  looksLikeMarkdown,
  markdownToRichTextHTML,
} from "@/lib/tiptap/markdown";
import type { ImportStoryPayload } from "./api";
import type { ImportPerson, ImportTask, ImportTeam } from "./schema";
import { isValidImportTaskEffort } from "./schema";

const STATUS_CATEGORY_ALIASES: {
  category: StateCategory;
  matches: readonly string[];
}[] = [
  {
    category: "completed",
    matches: ["done", "closed", "complete", "completed", "resolved"],
  },
  {
    category: "cancelled",
    matches: ["cancelled", "canceled", "rejected", "won't do", "wont do"],
  },
  {
    category: "paused",
    matches: ["paused", "blocked", "on hold", "waiting"],
  },
  {
    category: "started",
    matches: ["in progress", "doing", "started", "active", "review"],
  },
  { category: "backlog", matches: ["backlog", "icebox"] },
  {
    category: "unstarted",
    matches: ["to do", "todo", "open", "new", "ready", "selected"],
  },
];

const IMPORT_DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;
const IMPORT_HEX_COLOR_PATTERN = /^#[0-9A-Fa-f]{6}$/;
const IMPORT_TEAM_CODE_CHARACTERS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
const IMPORT_TEAM_COLORS = [
  "#4A90E2",
  "#7C5CFC",
  "#D85BBE",
  "#E05D5D",
  "#E08A32",
  "#2F9E80",
  "#2E9BC6",
  "#697386",
] as const;

const normalizeComparableText = (value: string) =>
  value.normalize("NFKC").trim().toLowerCase().replace(/\s+/g, " ");

const normalizeStatusText = (value: string) =>
  normalizeComparableText(value).replace(/[_-]+/g, " ");

const normalizeEmail = (value: string) =>
  value.normalize("NFKC").trim().toLowerCase();

const hasControlCharacter = (value: string) => /\p{Cc}/u.test(value);

const toHex = (bytes: ArrayBuffer) =>
  Array.from(new Uint8Array(bytes), (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("");

export const getImportParentCycleSourceIds = (
  tasks: readonly ImportTask[],
  canLinkParent: (task: ImportTask, parent: ImportTask) => boolean = () => true,
) => {
  const sourceIdCounts = new Map<string, number>();
  for (const task of tasks) {
    sourceIdCounts.set(
      task.sourceId,
      (sourceIdCounts.get(task.sourceId) ?? 0) + 1,
    );
  }
  const uniqueTasks = new Map(
    tasks
      .filter((task) => sourceIdCounts.get(task.sourceId) === 1)
      .map((task) => [task.sourceId, task]),
  );
  const parentBySourceId = new Map<string, string>();
  for (const task of uniqueTasks.values()) {
    if (!task.parentSourceId) continue;
    const parent = uniqueTasks.get(task.parentSourceId);
    if (parent && canLinkParent(task, parent)) {
      parentBySourceId.set(task.sourceId, parent.sourceId);
    }
  }

  const completed = new Set<string>();
  const cycleSourceIds = new Set<string>();
  for (const sourceId of uniqueTasks.keys()) {
    if (completed.has(sourceId)) continue;
    const path: string[] = [];
    const pathIndexes = new Map<string, number>();
    let currentSourceId: string | undefined = sourceId;
    while (
      currentSourceId &&
      !completed.has(currentSourceId) &&
      !pathIndexes.has(currentSourceId)
    ) {
      pathIndexes.set(currentSourceId, path.length);
      path.push(currentSourceId);
      currentSourceId = parentBySourceId.get(currentSourceId);
    }
    if (currentSourceId && pathIndexes.has(currentSourceId)) {
      const cycleStart = pathIndexes.get(currentSourceId) ?? path.length;
      for (const cycleSourceId of path.slice(cycleStart)) {
        cycleSourceIds.add(cycleSourceId);
      }
    }
    for (const visitedSourceId of path) completed.add(visitedSourceId);
  }
  return cycleSourceIds;
};

export const getBoundedImportSourceKey = async (sourceId: string) => {
  const value = sourceId.trim();
  const encoded = new TextEncoder().encode(value);
  if (value && encoded.byteLength <= 256 && !hasControlCharacter(value)) {
    return value;
  }

  const digest = await globalThis.crypto.subtle.digest("SHA-256", encoded);
  return `source:${toHex(digest)}`;
};

const getInferredCategory = (sourceStatus: string) => {
  const value = normalizeStatusText(sourceStatus);
  return STATUS_CATEGORY_ALIASES.find(({ matches }) =>
    matches.some((match) => value === match || value.startsWith(`${match} `)),
  )?.category;
};

type ImportStatusCandidate = Pick<
  State,
  "category" | "id" | "isDefault" | "name" | "orderIndex"
>;

const getStatusForCategory = <T extends ImportStatusCandidate>(
  category: StateCategory,
  statuses: T[],
): T | undefined => {
  const categoryStatuses = statuses.filter(
    (status) => status.category === category,
  );
  return (
    categoryStatuses.find((status) => status.isDefault) ??
    [...categoryStatuses]
      .sort((left, right) => left.orderIndex - right.orderIndex)
      .at(0)
  );
};

export const resolveImportStatus = <T extends ImportStatusCandidate>(
  sourceStatus: string | null,
  statuses: T[],
  sourceCategory?: StateCategory | null,
): T | undefined => {
  const category =
    sourceCategory ?? (sourceStatus ? getInferredCategory(sourceStatus) : null);
  if (sourceStatus) {
    const exactMatches = statuses.filter(
      (status) =>
        normalizeComparableText(status.name) ===
        normalizeComparableText(sourceStatus),
    );
    if (exactMatches.length === 1) return exactMatches[0];
    if (exactMatches.length > 1 && category) {
      const categorizedExactMatches = exactMatches.filter(
        (status) => status.category === category,
      );
      if (categorizedExactMatches.length === 1) {
        return categorizedExactMatches[0];
      }
      if (categorizedExactMatches.length > 1) {
        return getStatusForCategory(category, categorizedExactMatches);
      }
    }
  }

  if (category) {
    const categoryStatus = getStatusForCategory(category, statuses);
    if (categoryStatus) return categoryStatus;
  }

  return (
    statuses.find(
      (status) => status.isDefault && status.category === "unstarted",
    ) ??
    statuses.find((status) => status.isDefault) ??
    [...statuses].sort((a, b) => a.orderIndex - b.orderIndex).at(0)
  );
};

export type ImportMemberResolution = {
  member: Member;
  matchedBy: "email" | "fullName" | "username";
  requiresReview: boolean;
};

export type ImportMemberSuggestion = {
  member: Member;
  matchedBy: "fullName" | "username";
  score: number;
};

export type ImportEntityNameMatch<T> =
  | { kind: "none" }
  | { kind: "unique"; entity: T }
  | { kind: "ambiguous"; entities: T[] };

const getEligibleImportMembers = (members: Member[]) =>
  members.filter((member) => member.isActive && !member.isSystem);

const getEditDistance = (left: string, right: string) => {
  if (left === right) return 0;
  if (!left) return right.length;
  if (!right) return left.length;

  let previous = Array.from({ length: right.length + 1 }, (_, index) => index);
  for (let leftIndex = 1; leftIndex <= left.length; leftIndex += 1) {
    const current = [leftIndex];
    for (let rightIndex = 1; rightIndex <= right.length; rightIndex += 1) {
      const substitutionCost =
        left[leftIndex - 1] === right[rightIndex - 1] ? 0 : 1;
      current[rightIndex] = Math.min(
        (current[rightIndex - 1] ?? 0) + 1,
        (previous[rightIndex] ?? 0) + 1,
        (previous[rightIndex - 1] ?? 0) + substitutionCost,
      );
    }
    previous = current;
  }
  return previous[right.length] ?? Math.max(left.length, right.length);
};

const getImportNameSimilarity = (sourceName: string, candidateName: string) => {
  const source = normalizeComparableText(sourceName);
  const candidate = normalizeComparableText(candidateName);
  if (!source || !candidate) return 0;
  if (source === candidate) return 1;

  const editSimilarity =
    1 -
    getEditDistance(source, candidate) /
      Math.max(source.length, candidate.length);
  const sourceTokens = source.split(" ");
  const candidateTokens = candidate.split(" ");
  const availableCandidateIndexes = new Set(
    candidateTokens.map((_, index) => index),
  );
  let tokenScore = 0;
  for (const sourceToken of sourceTokens) {
    let bestIndex = -1;
    let bestScore = 0;
    for (const candidateIndex of availableCandidateIndexes) {
      const candidateToken = candidateTokens[candidateIndex] ?? "";
      let score = 0;
      if (sourceToken === candidateToken) {
        score = 1;
      } else if (
        sourceToken.startsWith(candidateToken) ||
        candidateToken.startsWith(sourceToken)
      ) {
        score =
          Math.min(sourceToken.length, candidateToken.length) /
          Math.max(sourceToken.length, candidateToken.length);
      }
      if (score > bestScore) {
        bestIndex = candidateIndex;
        bestScore = score;
      }
    }
    if (bestIndex >= 0) availableCandidateIndexes.delete(bestIndex);
    tokenScore += bestScore;
  }
  const tokenSimilarity =
    tokenScore / Math.max(sourceTokens.length, candidateTokens.length);
  return Math.max(editSimilarity, tokenSimilarity);
};

export const suggestImportPersonMember = (
  person: Pick<ImportPerson, "name">,
  members: Member[],
): ImportMemberSuggestion | undefined => {
  const sourceName = person.name?.trim();
  if (!sourceName) return undefined;

  const suggestions = getEligibleImportMembers(members)
    .map((member) => {
      const fullNameScore = getImportNameSimilarity(
        sourceName,
        member.fullName,
      );
      const usernameScore = getImportNameSimilarity(
        sourceName,
        member.username,
      );
      return usernameScore > fullNameScore
        ? { matchedBy: "username" as const, member, score: usernameScore }
        : { matchedBy: "fullName" as const, member, score: fullNameScore };
    })
    .sort((left, right) => right.score - left.score);
  const best = suggestions.at(0);
  const runnerUp = suggestions.at(1);
  if (!best || best.score < 0.68) return undefined;
  if (runnerUp && best.score - runnerUp.score < 0.08) return undefined;
  return best;
};

export const resolveImportPerson = (
  person: Pick<ImportPerson, "email" | "name">,
  members: Member[],
): ImportMemberResolution | undefined => {
  const eligibleMembers = getEligibleImportMembers(members);
  const email = person.email ? normalizeEmail(person.email) : "";

  if (email) {
    const emailMatches = eligibleMembers.filter(
      (member) => normalizeEmail(member.email) === email,
    );
    if (emailMatches.length === 1) {
      return {
        member: emailMatches[0],
        matchedBy: "email",
        requiresReview: false,
      };
    }
    return undefined;
  }

  const name = normalizeComparableText(person.name ?? "");
  if (!name) return undefined;

  const nameMatches = new Map<
    string,
    { member: Member; matchedBy: "fullName" | "username" }
  >();
  for (const member of eligibleMembers) {
    const matchesFullName = normalizeComparableText(member.fullName) === name;
    const matchesUsername = normalizeComparableText(member.username) === name;
    if (!matchesFullName && !matchesUsername) continue;
    nameMatches.set(member.id, {
      member,
      matchedBy: matchesFullName ? "fullName" : "username",
    });
  }

  if (nameMatches.size !== 1) return undefined;
  const match = nameMatches.values().next().value;
  if (!match) return undefined;
  return { ...match, requiresReview: true };
};

export const getImportPersonIdentityKey = (
  person: Pick<ImportPerson, "email" | "name"> | undefined,
  sourceId?: string | null,
) => {
  const email = person?.email ? normalizeEmail(person.email) : "";
  if (email) return `email:${email}`;

  const name = normalizeComparableText(person?.name ?? "");
  if (name) return `name:${name}`;

  const normalizedSourceId = sourceId?.trim();
  return normalizedSourceId ? `source:${normalizedSourceId}` : undefined;
};

export const getImportPersonSourceIdentityKey = (
  person: Pick<ImportPerson, "email" | "name"> | undefined,
  sourceId?: string | null,
) => {
  const normalizedSourceId = sourceId?.trim();
  if (normalizedSourceId) return `source:${normalizedSourceId}`;
  return getImportPersonIdentityKey(person);
};

export type ImportPersonIdentityUse = {
  identity: Pick<ImportPerson, "email" | "name"> | undefined;
  sourceId?: string | null;
};

export const analyzeImportPersonIdentityClaims = (
  identities: readonly ImportPersonIdentityUse[],
) => {
  const claimsByIdentityKey = new Map<
    string,
    { emails: Map<string, string>; names: Map<string, string> }
  >();
  for (const { identity, sourceId } of identities) {
    const identityKey = getImportPersonSourceIdentityKey(identity, sourceId);
    if (!identityKey) continue;
    const claims = claimsByIdentityKey.get(identityKey) ?? {
      emails: new Map<string, string>(),
      names: new Map<string, string>(),
    };
    const email = identity?.email?.trim();
    if (email) claims.emails.set(normalizeEmail(email), email);
    const name = identity?.name?.trim();
    if (name) claims.names.set(normalizeComparableText(name), name);
    claimsByIdentityKey.set(identityKey, claims);
  }

  const canonicalIdentities = new Map<
    string,
    { email: string | null; name: string | null }
  >();
  const conflictedIdentityKeys = new Set<string>();
  for (const [identityKey, claims] of claimsByIdentityKey) {
    if (
      claims.emails.size > 1 ||
      (claims.emails.size === 0 && claims.names.size > 1)
    ) {
      conflictedIdentityKeys.add(identityKey);
      continue;
    }
    canonicalIdentities.set(identityKey, {
      email: claims.emails.values().next().value ?? null,
      name: claims.names.values().next().value ?? null,
    });
  }
  return { canonicalIdentities, conflictedIdentityKeys };
};

export const getAmbiguousImportPersonNameIdentityKeys = (
  identities: readonly ImportPersonIdentityUse[],
) => {
  const identityKeysByName = new Map<string, Set<string>>();
  for (const { identity, sourceId } of identities) {
    const name = normalizeComparableText(identity?.name ?? "");
    const identityKey = getImportPersonSourceIdentityKey(identity, sourceId);
    if (!name || !identityKey) continue;
    const identityKeys = identityKeysByName.get(name) ?? new Set<string>();
    identityKeys.add(identityKey);
    identityKeysByName.set(name, identityKeys);
  }

  const ambiguousIdentityKeys = new Set<string>();
  for (const identityKeys of identityKeysByName.values()) {
    if (identityKeys.size < 2) continue;
    for (const identityKey of identityKeys) {
      ambiguousIdentityKeys.add(identityKey);
    }
  }
  return ambiguousIdentityKeys;
};

export const resolveImportAssignee = (
  assigneeEmail: string | null,
  members: Member[],
  assigneeName: string | null = null,
) => {
  const resolution = resolveImportPerson(
    { email: assigneeEmail, name: assigneeName ?? "" },
    members,
  );
  return resolution?.requiresReview ? undefined : resolution?.member;
};

export const resolveImportEntityByName = <
  T extends { name: string; teamId: string | null },
>(
  sourceName: string,
  teamId: string,
  entities: T[],
): T | undefined => {
  const resolution = resolveImportEntityNameMatch(
    sourceName,
    entities.filter((entity) => entity.teamId === teamId),
  );
  return resolution.kind === "unique" ? resolution.entity : undefined;
};

export const resolveImportEntityNameMatch = <T extends { name: string }>(
  sourceName: string,
  entities: T[],
): ImportEntityNameMatch<T> => {
  const normalizedName = normalizeComparableText(sourceName);
  if (!normalizedName) return { kind: "none" };

  const matches = entities.filter(
    (entity) => normalizeComparableText(entity.name) === normalizedName,
  );
  if (matches.length === 0) return { kind: "none" };
  if (matches.length === 1) return { kind: "unique", entity: matches[0] };
  return { kind: "ambiguous", entities: matches };
};

export const isValidImportDate = (
  value: string | null | undefined,
): value is string => {
  if (!value) return false;
  if (!IMPORT_DATE_PATTERN.test(value)) return false;

  const year = Number(value.slice(0, 4));
  const month = Number(value.slice(5, 7));
  const day = Number(value.slice(8, 10));
  const date = new Date(Date.UTC(year, month - 1, day));
  return (
    date.getUTCFullYear() === year &&
    date.getUTCMonth() === month - 1 &&
    date.getUTCDate() === day
  );
};

export const isValidImportDateRange = (
  startDate: string | null | undefined,
  endDate: string | null | undefined,
) => {
  const hasStartDate = startDate !== null && startDate !== undefined;
  const hasEndDate = endDate !== null && endDate !== undefined;
  if (hasStartDate && !isValidImportDate(startDate)) return false;
  if (hasEndDate && !isValidImportDate(endDate)) return false;
  return !hasStartDate || !hasEndDate || endDate >= startDate;
};

const importStringHash = (value: string) => {
  let hash = 0;
  for (let index = 0; index < value.length; index++) {
    hash = (hash * 31 + value.charCodeAt(index)) % 2_147_483_647;
  }
  return hash;
};

const sanitizeImportTeamCode = (value: string) =>
  value
    .normalize("NFKC")
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, "");

const baseImportTeamCode = (team: Pick<ImportTeam, "code" | "name">) => {
  const requested = sanitizeImportTeamCode(team.code ?? "");
  if (requested.length >= 2) return requested.slice(0, 3);

  const words = team.name
    .normalize("NFKC")
    .toUpperCase()
    .split(/[^A-Z0-9]+/)
    .filter(Boolean);
  const initials = words.map((word) => word[0]).join("");
  const compact = sanitizeImportTeamCode(team.name);
  const candidate = initials.length >= 2 ? initials : compact;
  if (candidate.length >= 2) return candidate.slice(0, 3);
  if (candidate.length === 1) return `${candidate}X`;
  return "TM";
};

export const deriveImportTeamCode = (
  team: Pick<ImportTeam, "code" | "name" | "sourceId">,
  unavailableCodes: Iterable<string> = [],
) => {
  const unavailable = new Set(
    Array.from(unavailableCodes, (code) => sanitizeImportTeamCode(code)),
  );
  const base = baseImportTeamCode(team);
  if (!unavailable.has(base)) return base;

  const hash = importStringHash(
    `${team.sourceId}\u0000${team.name}\u0000${team.code ?? ""}`,
  );
  const stem = base.slice(0, 2).padEnd(2, "X");
  for (let offset = 0; offset < IMPORT_TEAM_CODE_CHARACTERS.length; offset++) {
    const suffix =
      IMPORT_TEAM_CODE_CHARACTERS[
        (hash + offset) % IMPORT_TEAM_CODE_CHARACTERS.length
      ];
    const candidate = `${stem}${suffix}`;
    if (!unavailable.has(candidate)) return candidate;
  }
  throw new Error(`Unable to derive a unique code for team ${team.name}`);
};

export const deriveImportTeamColor = (
  team: Pick<ImportTeam, "color" | "name" | "sourceId">,
) => {
  if (team.color && IMPORT_HEX_COLOR_PATTERN.test(team.color)) {
    return team.color.toUpperCase();
  }
  const hash = importStringHash(`${team.sourceId}\u0000${team.name}`);
  return IMPORT_TEAM_COLORS[hash % IMPORT_TEAM_COLORS.length];
};

export const toImportStoryPayload = ({
  allowAutomaticAssigneeResolution = true,
  assigneeId,
  keyResultId,
  labelIds,
  members,
  objectiveId,
  parentId,
  sprintId,
  statuses,
  task,
  teamId,
}: {
  allowAutomaticAssigneeResolution?: boolean;
  assigneeId?: string;
  keyResultId?: string;
  labelIds?: string[];
  members: Member[];
  objectiveId?: string;
  parentId?: string;
  sprintId?: string;
  statuses: State[];
  task: ImportTask;
  teamId: string;
}): ImportStoryPayload => {
  if (!isValidImportDateRange(task.startDate, task.endDate)) {
    throw new Error(`Imported task ${task.sourceId} has an invalid date range`);
  }
  if (!isValidImportTaskEffort(task)) {
    throw new Error(`Imported task ${task.sourceId} has invalid effort values`);
  }

  const status = resolveImportStatus(
    task.status,
    statuses,
    task.statusCategory,
  );
  const assignee =
    !assigneeId && allowAutomaticAssigneeResolution
      ? resolveImportAssignee(task.assigneeEmail, members, task.assigneeName)
      : undefined;
  const resolvedAssigneeId = assigneeId ?? assignee?.id;
  const resolvedLabelIds = labelIds
    ? [...new Set(labelIds.filter(Boolean))]
    : undefined;
  const description = task.description.trim();
  const descriptionHTML = looksLikeMarkdown(description)
    ? markdownToRichTextHTML(description)
    : undefined;

  return {
    title: task.title.trim(),
    description,
    ...(descriptionHTML ? { descriptionHTML } : {}),
    teamId,
    ...(status ? { statusId: status.id } : {}),
    ...(resolvedAssigneeId ? { assigneeId: resolvedAssigneeId } : {}),
    ...(objectiveId ? { objectiveId } : {}),
    ...(keyResultId ? { keyResultId } : {}),
    ...(sprintId ? { sprintId } : {}),
    ...(parentId ? { parentId } : {}),
    ...(resolvedLabelIds?.length ? { labelIds: resolvedLabelIds } : {}),
    priority: task.priority,
    ...(typeof task.estimateValue === "number"
      ? { estimateValue: task.estimateValue }
      : {}),
    ...(typeof task.estimatedDurationMinutes === "number"
      ? { estimatedDurationMinutes: task.estimatedDurationMinutes }
      : {}),
    ...(typeof task.minimumFocusBlockMinutes === "number"
      ? { minimumFocusBlockMinutes: task.minimumFocusBlockMinutes }
      : {}),
    ...(task.startDate ? { startDate: task.startDate } : {}),
    ...(task.endDate ? { endDate: task.endDate } : {}),
  };
};
