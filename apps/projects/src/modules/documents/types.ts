export type DocumentVisibility = "workspace" | "restricted" | "private";
export type DocumentRelationType = "story" | "objective";
export type DocumentScope = "all" | "mine" | "shared" | "templates";

export type DocumentMember = {
  userId: string;
  role: "viewer" | "editor";
};

export type RelatedWork = {
  entityId: string;
  entityType: DocumentRelationType;
  title: string;
  reference: string | null;
  teamId: string | null;
};

export type WorkspaceDocument = {
  id: string;
  workspaceId: string;
  title: string;
  contentHtml: string;
  contentText: string;
  visibility: DocumentVisibility;
  createdBy: string;
  updatedBy: string;
  createdAt: string;
  updatedAt: string;
  canEdit: boolean;
  sharedWith: DocumentMember[];
  relatedWork: RelatedWork[];
  relatedWorkCount: number;
};

export type WorkspaceDocumentSummary = Pick<
  WorkspaceDocument,
  | "id"
  | "workspaceId"
  | "title"
  | "visibility"
  | "createdBy"
  | "updatedBy"
  | "createdAt"
  | "updatedAt"
  | "canEdit"
  | "relatedWorkCount"
>;

export type DocumentAccessFilter = DocumentVisibility | "all";
export type DocumentOwnerFilter = "all" | "mine" | "others";
export type DocumentUpdatedFilter = "all" | "today" | "7d" | "30d" | "90d";
export type DocumentSortField = "updated" | "title";
export type DocumentSortDirection = "asc" | "desc";

export type DocumentListState = {
  access: DocumentAccessFilter;
  direction: DocumentSortDirection;
  owner: DocumentOwnerFilter;
  page: number;
  sort: DocumentSortField;
  updated: DocumentUpdatedFilter;
};

export type DocumentCreate = {
  title?: string;
  contentHtml?: string;
  contentText?: string;
  visibility?: DocumentVisibility;
};

export type DocumentUpdate = Partial<
  Pick<WorkspaceDocument, "title" | "contentHtml" | "contentText">
>;

export type DocumentAccessUpdate = {
  visibility: DocumentVisibility;
  members: DocumentMember[];
};

export type DocumentMedia = {
  id: string;
  filename: string;
  size: number;
  mimeType: string;
  url: string;
  createdAt: string;
  uploadedBy: string;
};
