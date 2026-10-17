import type { Editor } from "@tiptap/core";
import { Box, Button, Flex, Text } from "ui";
import { toast } from "sonner";
import type { useDocumentCollaboration } from "./use-document-collaboration";

type DocumentSyncNoticeProps = {
  collaboration: ReturnType<typeof useDocumentCollaboration>;
  collaborative: boolean;
  hasSaveError: boolean;
  isDesktopApp: boolean;
  editor: Editor | null;
  onOpenRestored: () => void;
};

export function DocumentSyncNotice({
  collaboration,
  collaborative,
  hasSaveError,
  isDesktopApp,
  editor,
  onOpenRestored,
}: DocumentSyncNoticeProps) {
  return (
    <>
      {(collaboration.configured &&
        (!isDesktopApp ||
          !collaboration.writable ||
          Boolean(collaboration.storageError)) &&
        (collaboration.status === "blocked" ||
          collaboration.status === "offline" ||
          Boolean(collaboration.error) ||
          Boolean(collaboration.storageError))) ||
      hasSaveError ||
      (!collaboration.configured && collaborative) ? (
        <Box
          className="bg-surface-elevated border-border mb-6 rounded-xl border p-4"
          role="status"
        >
          <Text>
            {collaboration.storageError ??
              collaboration.error ??
              (collaboration.writable
                ? "You can keep editing offline. Changes are saved on this device and merge when you reconnect."
                : "Editing is paused. Your connection or document access changed. Preserve any unsaved work before reloading.")}
          </Text>
          <Flex className="mt-3 gap-2">
            {collaboration.retry ? (
              <Button
                color="tertiary"
                onClick={() => {
                  void collaboration
                    .retry()
                    .catch((reason: unknown) => toast.error(String(reason)));
                }}
                size="sm"
              >
                Retry connection and save
              </Button>
            ) : (
              <Button
                color="tertiary"
                onClick={() => {
                  window.location.reload();
                }}
                size="sm"
              >
                Reload document
              </Button>
            )}
            {collaboration.exported &&
            collaboration.status === "blocked" &&
            collaboration.openRestored ? (
              <Button
                color="tertiary"
                onClick={() => {
                  onOpenRestored();
                }}
                size="sm"
              >
                Open restored version
              </Button>
            ) : null}
            {collaboration.exportRecovery ? (
              <Button
                color="tertiary"
                onClick={() => {
                  void collaboration
                    .exportRecovery({
                      contentHtml: editor?.getHTML() ?? "",
                      contentText: editor?.getText() ?? "",
                    })
                    .catch((reason: unknown) => toast.error(String(reason)));
                }}
                size="sm"
              >
                Export recovery file
              </Button>
            ) : null}
          </Flex>
        </Box>
      ) : null}
    </>
  );
}
