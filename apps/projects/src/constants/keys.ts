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
