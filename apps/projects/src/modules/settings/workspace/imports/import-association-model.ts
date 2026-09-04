import type { StoryAssociationType } from "@/shared/story/types";
import type { ImportTask } from "./schema";

export const getCanonicalImportAssociation = (
  sourceId: string,
  targetId: string,
  sourceType: ImportTask["associations"][number]["type"],
): {
  fromId: string;
  toId: string;
  type: StoryAssociationType;
} => {
  if (sourceType === "blocked_by") {
    return { fromId: targetId, toId: sourceId, type: "blocking" };
  }
  if (sourceType === "blocks") {
    return { fromId: sourceId, toId: targetId, type: "blocking" };
  }
  const [fromId, toId] =
    sourceId < targetId ? [sourceId, targetId] : [targetId, sourceId];
  return { fromId, toId, type: sourceType };
};

export const getImportAssociationKey = ({
  fromId,
  toId,
  type,
}: {
  fromId: string;
  toId: string;
  type: StoryAssociationType;
}) => {
  if (type === "blocking") return `${type}\u0000${fromId}\u0000${toId}`;
  const [firstId, secondId] = fromId < toId ? [fromId, toId] : [toId, fromId];
  return `${type}\u0000${firstId}\u0000${secondId}`;
};
