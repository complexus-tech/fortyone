export type ApiResponse<T> = {
  data?: T;
  error?: {
    message: string;
  };
};

export type Member = {
  id: string | null;
  username: string;
  email: string;
  fullName: string;
  avatarUrl: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export type Label = {
  id: string;
  name: string;
  color: string;
  teamId: string | null;
  workspaceId: string;
  createdAt: string;
  updatedAt: string;
};

export type Link = {
  id: string;
  url: string;
  title: string;
  storyId: string;
  createdAt: string;
  updatedAt: string;
};

export type Comment = {
  id: string;
  storyId: string;
  parentId: string | null;
  userId: string;
  comment: string;
  createdAt: string;
  updatedAt: string;
  subComments: Comment[];
};

export type StoriesSummary = {
  closed: number;
  overdue: number;
  inProgress: number;
  created: number;
  assigned: number;
};

export type Workspace = {
  id: string;
  name: string;
  createdAt: string;
  updatedAt: string;
};
