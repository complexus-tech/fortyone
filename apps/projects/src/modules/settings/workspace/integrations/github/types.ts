export type GitHubWorkspaceSettings = {
  branchFormat: string;
  linkCommitsByMagicWords: boolean;
  syncAssignees: boolean;
  syncLabels: boolean;
  autoPopulatePrBody: boolean;
  closeOnCommitKeywords: boolean;
  createdAt: string;
  updatedAt: string;
};

export type GitHubInstallation = {
  id: string;
  githubInstallationId: number;
  accountLogin: string;
  accountType: string;
  accountAvatarUrl?: string | null;
  repositorySelection: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export type GitHubRepository = {
  id: string;
  installationId: string;
  githubRepositoryId: number;
  ownerLogin: string;
  name: string;
  fullName: string;
  description?: string | null;
  htmlUrl: string;
  defaultBranch: string;
  isPrivate: boolean;
  isArchived: boolean;
  isDisabled: boolean;
  isActive: boolean;
  lastSyncedAt?: string | null;
};

export type GitHubIssueSyncLink = {
  id: string;
  repositoryId: string;
  repositoryName: string;
  teamId: string;
  teamName: string;
  teamColor: string;
  syncDirection: "inbound_only" | "bidirectional";
  isActive: boolean;
};

export type GitHubWorkflowRule = {
  id: string;
  eventKey: string;
  targetStatusId?: string | null;
  baseBranchPattern?: string | null;
  isActive: boolean;
};

export type GitHubIntegration = {
  settings: GitHubWorkspaceSettings;
  installations: GitHubInstallation[];
  repositories: GitHubRepository[];
  issueSyncLinks: GitHubIssueSyncLink[];
};

export type CreateGitHubInstallSessionResponse = {
  installUrl: string;
};

export type CreateGitHubIssueSyncLinkInput = {
  repositoryId: string;
  teamId: string;
  syncDirection: "inbound_only" | "bidirectional";
};

export type UpdateGitHubIssueSyncLinkInput = Partial<{
  syncDirection: "inbound_only" | "bidirectional";
  isActive: boolean;
}>;

export type UpdateGitHubWorkspaceSettingsInput = Partial<{
  branchFormat: string;
  linkCommitsByMagicWords: boolean;
  syncAssignees: boolean;
  syncLabels: boolean;
  autoPopulatePrBody: boolean;
  closeOnCommitKeywords: boolean;
}>;

export type GitHubTeamSettings = {
  teamId: string;
  rules: GitHubWorkflowRule[];
};

export type UpdateGitHubTeamSettingsInput = {
  rules: Array<{
    eventKey: string;
    targetStatusId?: string | null;
    baseBranchPattern?: string | null;
    isActive: boolean;
  }>;
};
