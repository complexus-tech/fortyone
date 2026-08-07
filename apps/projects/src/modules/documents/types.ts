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
