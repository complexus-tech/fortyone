import type { QueryClient } from "@tanstack/react-query";
import type { StoryAttachment } from "@/modules/story/types";
import { storyKeys } from "@/modules/stories/constants";
import type { DesktopCacheScope } from "./cache-types";
import type { DesktopOperation, DesktopOperationAck } from "./outbox-types";
import { setChanged } from "./outbox-records";

export function projectAttachment(
  client: QueryClient,
  scope: DesktopCacheScope,
  operation: DesktopOperation,
) {
  const { attachmentId, fileName, contentType, size, createdAt } =
    operation.payload;
  if (
    typeof attachmentId !== "string" ||
    typeof fileName !== "string" ||
    typeof contentType !== "string" ||
    typeof size !== "number" ||
    typeof createdAt !== "string"
  )
    return;
  const key = storyKeys.attachments(scope.workspaceSlug, operation.resourceId);
  const current = client.getQueryData<StoryAttachment[]>(key) ?? [];
  if (current.some((item) => item.id === attachmentId)) return;
  setChanged(client, key, [
    ...current,
    {
      id: attachmentId,
      filename: fileName,
      mimeType: contentType,
      size,
      createdAt,
      uploadedBy: scope.accountId,
      url: "",
      pending: true,
    },
  ]);
}

export function acknowledgeAttachment(
  client: QueryClient,
  scope: DesktopCacheScope,
  ack: DesktopOperationAck,
) {
  if (!ack.storyId) return;
  const key = storyKeys.attachments(scope.workspaceSlug, ack.storyId);
  const current = client.getQueryData<StoryAttachment[]>(key) ?? [];
  const server = current.find((item) => item.id === ack.resourceId);
  const provisional = current.find((item) => item.id === ack.provisionalId);
  if (!server && !provisional && !ack.resource) return;
  // A later server refetch owns its current URL and metadata.
  const attachment = server ?? {
    ...ack.resource,
    id: ack.resourceId,
    pending: false,
  };
  const next = current.filter(
    (item) => item.id !== ack.provisionalId && item.id !== ack.resourceId,
  );
  setChanged(client, key, [...next, attachment]);
}
