export type WorkspaceSettings = {
  storyTerm: "issue" | "task" | "story";
  sprintTerm: "sprint" | "cycle" | "iteration";
  objectiveTerm: "objective" | "goal" | "project";
  keyResultTerm: "key result" | "milestone" | "focus area";
  objectiveEnabled: boolean;
  keyResultEnabled: boolean;
};
