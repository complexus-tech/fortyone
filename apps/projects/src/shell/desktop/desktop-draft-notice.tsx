"use client";
import { Button, Flex, Text } from "ui";

export function DesktopDraftNotice({
  recovered,
  error,
  onRestore,
  onDiscard,
  onRetry,
}: {
  recovered: boolean;
  status: string;
  error: string | null;
  onRestore: () => void;
  onDiscard: () => void;
  onRetry: () => void;
}) {
  if (!recovered && !error) return null;
  return (
    <Flex align="center" className="my-2 flex-wrap" gap={2}>
      <Text color="muted" role={error ? "alert" : "status"}>
        {error ?? "A saved draft is available on this device."}
      </Text>
      {recovered ? (
        <>
          <Button onClick={onRestore} size="sm" variant="naked">
            Restore draft
          </Button>
          <Button onClick={onDiscard} size="sm" variant="naked">
            Discard saved draft
          </Button>
        </>
      ) : null}
      {error ? (
        <Button onClick={onRetry} size="sm" variant="naked">
          Retry saving
        </Button>
      ) : null}
    </Flex>
  );
}
