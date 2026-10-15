export type DesktopCacheScope = {
  accountId: string;
  workspaceId: string;
  workspaceSlug: string;
};

export type DesktopCachedQuery = {
  queryKey: readonly unknown[];
  data: unknown;
  updatedAt: number;
};

export type DesktopCacheEnvelope = DesktopCacheScope & {
  version: 1;
  savedAt: number;
  queries: DesktopCachedQuery[];
};
