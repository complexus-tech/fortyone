export type DocumentVisibility = "workspace" | "restricted" | "private";
export type DocumentRelationType = "story" | "objective";
export type DocumentScope =
  | "all"
  | "mine"
  | "shared"
  | "favorites"
  | "archived"
  | "templates";

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
  statusColor?: string | null;
  referencedInContent?: boolean;
};

export type WorkspaceDocument = {
  revision: number;
  collaborationEpoch: number;
  collaborative: boolean;
  publicToken: string | null;
  publicEnabled?: boolean;
  archivedAt?: string | null;
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
> & {
  publicEnabled?: boolean;
  archivedAt?: string | null;
  relatedWork?: RelatedWork[];
};

export type DocumentAccessFilter = DocumentVisibility | "public" | "all";
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

export type DocumentUpdate = { expectedRevision: number } & Partial<
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

export type DocumentRevision = {
  revision: number;
  title: string;
  contentHtml?: string;
  contentText?: string;
  editedBy: string | null;
  createdAt: string;
};
export type CollaborationSession = { token: string; name: string };

export type DocumentComment = {
  id: string;
  body: string;
  createdBy: string;
  authorName: string;
  authorAvatar: string | null;
  createdAt: string;
  updatedAt: string;
  replyToCommentId?: string | null;
};

export type DocumentCommentThread = {
  id: string;
  documentId: string;
  quote: string;
  anchorStart: number;
  anchorEnd: number;
  createdBy: string;
  resolvedAt: string | null;
  resolvedBy: string | null;
  createdAt: string;
  comments: DocumentComment[];
};

export type DocumentCommentSelection = {
  draftKey?: string;
  isDetached?: boolean;
  from: number;
  text: string;
  to: number;
};
