export const isCommentUnavailableError = (error: unknown) =>
  typeof error === "object" &&
  error !== null &&
  "status" in error &&
  [401, 403, 404, 410].includes(Number(error.status));
