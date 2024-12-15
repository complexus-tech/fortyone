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
