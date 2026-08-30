import { z } from "zod";

export const syncDirectionSchema = z.enum(["inbound_only", "bidirectional"]);

export const teamRuleSchema = z.object({
  baseBranchPattern: z
    .string()
    .nullable()
    .optional()
    .describe("Optional base branch pattern for pull request rules."),
  eventKey: z
    .string()
    .describe(
      "GitHub workflow event key, for example issue.opened, issue.closed, pull_request.opened.",
    ),
  isActive: z.boolean().describe("Whether the automation rule is active."),
  targetStatusId: z
    .string()
    .nullable()
    .optional()
    .describe("Target FortyOne status ID. Resolve statuses first."),
});

export const workspaceSettingsUpdateSchema = z.object({
  autoPopulatePrBody: z.boolean().optional(),
  branchFormat: z.string().optional(),
  closeOnCommitKeywords: z.boolean().optional(),
  linkCommitsByMagicWords: z.boolean().optional(),
  syncAssignees: z.boolean().optional(),
  syncLabels: z.boolean().optional(),
});

export type GitHubRepository = {
  defaultBranch: string;
  fullName: string;
  htmlUrl: string;
  id: string;
  isActive: boolean;
  isArchived: boolean;
  isDisabled: boolean;
  isPrivate: boolean;
  lastSyncedAt?: string | null;
  name: string;
  ownerLogin: string;
};

export type GitHubToolTeam = {
  code: string;
  color: string;
  id: string;
  name: string;
};

export type GitHubToolStory = {
  id: string;
  sequenceId: number;
  teamCode: string;
  title: string;
};

export type StoryGitHubLink = {
  checkState: string | null;
  createdAt: string;
  externalType: "branch" | "commit" | "issue" | "pull_request";
  githubNumber: number | null;
  id: string;
  refName: string | null;
  repositoryFullName: string;
  reviewState: string | null;
  state: string | null;
  title: string | null;
  url: string;
};

export const toRepositorySummary = (repository: GitHubRepository) => ({
  defaultBranch: repository.defaultBranch,
  fullName: repository.fullName,
  htmlUrl: repository.htmlUrl,
  id: repository.id,
  isActive: repository.isActive,
  isArchived: repository.isArchived,
  isDisabled: repository.isDisabled,
  isPrivate: repository.isPrivate,
  lastSyncedAt: repository.lastSyncedAt,
  name: repository.name,
  ownerLogin: repository.ownerLogin,
});

export const toStoryLinkSummary = (link: StoryGitHubLink) => ({
  checkState: link.checkState,
  createdAt: link.createdAt,
  id: link.id,
  number: link.githubNumber,
  refName: link.refName,
  repositoryFullName: link.repositoryFullName,
  reviewState: link.reviewState,
  state: link.state,
  title: link.title,
  type: link.externalType,
  url: link.url,
});

export const toWorkspaceSettingsUpdate = (
  input: z.infer<typeof workspaceSettingsUpdateSchema>,
) => ({
  ...(input.autoPopulatePrBody !== undefined
    ? { autoPopulatePrBody: input.autoPopulatePrBody }
    : {}),
  ...(input.branchFormat !== undefined
    ? { branchFormat: input.branchFormat }
    : {}),
  ...(input.closeOnCommitKeywords !== undefined
    ? { closeOnCommitKeywords: input.closeOnCommitKeywords }
    : {}),
  ...(input.linkCommitsByMagicWords !== undefined
    ? { linkCommitsByMagicWords: input.linkCommitsByMagicWords }
    : {}),
  ...(input.syncAssignees !== undefined
    ? { syncAssignees: input.syncAssignees }
    : {}),
  ...(input.syncLabels !== undefined ? { syncLabels: input.syncLabels } : {}),
});
