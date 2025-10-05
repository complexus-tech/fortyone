export type Workspace = {
  id: string;
  name: string;
  slug: string;
  color: string;
  avatarUrl: string | null;
  userRole: "admin" | "member" | "guest" | "system";
  trialEndsOn: string | null;
  deletedAt: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export type WorkspaceSettings = {
  storyTerm: "issue" | "task" | "story";
  sprintTerm: "sprint" | "cycle" | "iteration";
  objectiveTerm: "objective" | "goal" | "project";
  keyResultTerm: "key result" | "milestone" | "focus area";
  objectiveEnabled: boolean;
  keyResultEnabled: boolean;
};
