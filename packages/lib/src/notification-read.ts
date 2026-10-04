export type ReadNotificationInput = {
  id: string;
  observedCreatedAt?: string;
};

export const getNotificationReadPath = (
  id: string,
  observedCreatedAt?: string,
): string => {
  const path = `notifications/${encodeURIComponent(id)}/read`;
  if (!observedCreatedAt) return path;
  return `${path}?${new URLSearchParams({ observedCreatedAt }).toString()}`;
};
