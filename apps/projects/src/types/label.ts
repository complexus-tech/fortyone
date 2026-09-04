export type Label = {
  id: string;
  name: string;
  color: string;
  teamId: string | null;
  workspaceId: string;
  createdAt: string;
  updatedAt: string;
};

export type LabelsPage = {
  labels: Label[];
  pagination: {
    page: number;
    pageSize: number;
    hasMore: boolean;
    nextPage: number;
  };
};
