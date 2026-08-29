export type FigmaConnection = {
  id: string;
  workspaceId: string;
  figmaUserId: string;
  email: string | null;
  handle: string | null;
  scopes: string[];
  expiresAt: string;
  connectedByUserId: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export type FigmaIntegration = {
  configured: boolean;
  connection: FigmaConnection | null;
};

export type FigmaArtifact = {
  fileKey: string;
  nodeId: string | null;
  originalUrl: string;
  canonicalUrl: string;
  fileName: string;
  nodeName: string | null;
  nodeType: string | null;
  thumbnailUrl: string | null;
  version: string | null;
  lastModified: string | null;
  textContent?: string[];
};

export type FigmaHandoffStatus = "READY_FOR_DEV" | "COMPLETED";
export type FigmaHandoffStatuses = Record<string, FigmaHandoffStatus>;

export type StoryFigmaLink = {
  id: string;
  workspaceId: string;
  storyId: string;
  createdByUserId: string;
  storyLinkId: string | null;
  artifact: FigmaArtifact;
  devStatus: "NONE" | "READY_FOR_DEV" | "COMPLETED" | null;
  devResourceId: string | null;
  lastSyncedAt: string;
  unavailableAt: string | null;
  createdAt: string;
  updatedAt: string;
};
