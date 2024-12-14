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
