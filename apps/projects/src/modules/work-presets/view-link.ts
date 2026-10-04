export const savedViewPath = (teamId: string, viewId: string) =>
  `/teams/${encodeURIComponent(teamId)}/stories?view=${encodeURIComponent(viewId)}`;
