export const statusKeys = {
  all: ["statuses"] as const,
  lists: () => [...statusKeys.all, "list"] as const,
  details: () => [...statusKeys.all, "detail"] as const,
  detail: (id: string) => [...statusKeys.details(), id] as const,
};

export const statusTags = {
  all: "statuses" as const,
  lists: () => `${statusTags.all}-list` as const,
  details: () => `${statusTags.all}-detail` as const,
  detail: (id: string) => `${statusTags.details()}-${id}` as const,
};

export const sprintTags = {
  all: "sprints" as const,
  lists: () => `${sprintTags.all}-list` as const,
  details: () => `${sprintTags.all}-detail` as const,
  detail: (id: string) => `${sprintTags.details()}-${id}` as const,
};

export const sprintKeys = {
  all: ["sprints"] as const,
  lists: () => [...sprintKeys.all, "list"] as const,
  details: () => [...sprintKeys.all, "detail"] as const,
  detail: (id: string) => [...sprintKeys.details(), id] as const,
};

export const memberKeys = {
  all: ["members"] as const,
  lists: () => [...memberKeys.all, "list"] as const,
  details: () => [...memberKeys.all, "detail"] as const,
  detail: (id: string) => [...memberKeys.details(), id] as const,
};

export const memberTags = {
  all: "members" as const,
  lists: () => `${memberTags.all}-list` as const,
  details: () => `${memberTags.all}-detail` as const,
  detail: (id: string) => `${memberTags.details()}-${id}` as const,
};

export const labelTags = {
  all: "labels" as const,
  lists: () => `${labelTags.all}-list` as const,
  details: () => `${labelTags.all}-detail` as const,
  detail: (id: string) => `${labelTags.details()}-${id}` as const,
  team: (teamId: string) => `${labelTags.all}-${teamId}` as const,
};

export const labelKeys = {
  all: ["labels"] as const,
  lists: () => [...labelKeys.all, "list"] as const,
  details: () => [...labelKeys.all, "detail"] as const,
  detail: (id: string) => [...labelKeys.details(), id] as const,
  team: (teamId: string) => [...labelKeys.all, teamId] as const,
};

export const linkKeys = {
  story: (storyId: string) => ["story-links", storyId] as const,
  metadata: (url: string) => ["link-metadata", url] as const,
};

export const linkTags = {
  story: (storyId: string) => `story-links-${storyId}` as const,
  metadata: (url: string) => `link-metadata-${url}` as const,
};

export const workspaceTags = {
  all: "workspaces" as const,
  lists: () => `${workspaceTags.all}-list` as const,
  detail: () => `${workspaceTags.all}-detail` as const,
};

export const workspaceKeys = {
  all: ["workspaces"] as const,
  lists: () => [...workspaceKeys.all, "list"] as const,
  detail: () => [...workspaceKeys.all, "detail"] as const,
};

export const teamTags = {
  all: "teams" as const,
  lists: () => `${teamTags.all}-list` as const,
  details: () => `${teamTags.all}-detail` as const,
  detail: (id: string) => `${teamTags.details()}-${id}` as const,
};

export const teamKeys = {
  all: ["teams"] as const,
  lists: () => [...teamKeys.all, "list"] as const,
  details: () => [...teamKeys.all, "detail"] as const,
  detail: (id: string) => [...teamKeys.details(), id] as const,
};
