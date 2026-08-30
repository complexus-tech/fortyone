export type ApiResponse<T> = {
  data?: T | null;
  error?: {
    code?: string;
    hint?: string;
    message: string;
  };
};
