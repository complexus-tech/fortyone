export type GoogleDriveTargetType =
  | "story"
  | "objective"
  | "document"
  | "comment";

export type GoogleDriveTarget = {
  id: string;
  type: GoogleDriveTargetType;
};

export type GoogleDriveConnectionStatus =
  | "connected"
  | "disconnected"
  | "reauthorization_required";

export type GoogleDriveIntegration = {
  configured: boolean;
  connected: boolean;
  email?: string;
  status: GoogleDriveConnectionStatus;
  requiresReauthorization: boolean;
};

export type GoogleDrivePickerSession = {
  accessToken: string;
  apiKey: string;
  appId: string;
  origin?: string;
};

export type GoogleDrivePickerFile = {
  id: string;
  mimeType?: string;
  name?: string;
  resourceKey?: string;
};

export type GoogleDriveFileAvailability =
  | "available"
  | "access_required"
  | "deleted"
  | "reauthorization_required";

export type GoogleDriveFileReference = {
  id: string;
  name: string;
  mimeType: string;
  webViewLink: string;
  modifiedTime?: string;
  connectionEmail?: string;
  availability: GoogleDriveFileAvailability;
  targetType: GoogleDriveTargetType;
  targetId: string;
  createdAt: string;
  updatedAt: string;
};

export type GoogleDriveFileType = "document" | "spreadsheet";

export type GoogleDriveImportVisibility = "workspace" | "private";

export type GoogleDriveImportResult = {
  documentId: string;
  sourceReferenceId: string;
};
