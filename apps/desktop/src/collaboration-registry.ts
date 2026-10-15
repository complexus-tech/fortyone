const activeEditors = new Set<string>();
export const collaborationKey = (workspaceId: string, documentId: string) =>
  `${workspaceId}:${documentId}`;
export function registerCollaborationEditor(
  workspaceId: string,
  documentId: string,
) {
  const key = collaborationKey(workspaceId, documentId);
  activeEditors.add(key);
  return () => {
    activeEditors.delete(key);
  };
}
export function hasCollaborationEditor(
  workspaceId: string,
  documentId: string,
) {
  return activeEditors.has(collaborationKey(workspaceId, documentId));
}
