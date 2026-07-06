import type { Label } from "@/types";

const normalizeLabelName = (name: string) => name.trim().toLowerCase();

export const canCreateLabelFromQuery = (query: string, labels: Label[]) => {
  const normalizedQuery = normalizeLabelName(query);
  if (!normalizedQuery) {
    return false;
  }

  return !labels.some(
    (label) => normalizeLabelName(label.name) === normalizedQuery,
  );
};
