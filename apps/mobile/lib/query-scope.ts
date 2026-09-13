export type QueryScope = {
  userId: string | null;
  workspace: string | null;
};

export const scopedQueryKey = (scope: QueryScope, resource: string) =>
  ["session", scope.userId, scope.workspace, resource] as const;

export const queryPersistencePrefix = (userId: string | null) =>
  `fortyone:queries:v2:${JSON.stringify(userId)}:`;

export const queryPersistenceKey = (scope: QueryScope) =>
  `${queryPersistencePrefix(scope.userId)}${JSON.stringify(scope.workspace)}`;
