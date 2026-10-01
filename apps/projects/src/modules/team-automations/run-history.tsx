"use client";

import { Box, Button, Dialog, Flex, Text } from "ui";
import { useAutomationRuns } from "./hooks";

export const AutomationRunHistory = ({
  id,
  name,
  onClose,
}: {
  id: string;
  name: string;
  onClose: () => void;
}) => {
  const query = useAutomationRuns(id);
  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      open
    >
      <Dialog.Content
        className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]"
        size="md"
      >
        <Dialog.Header className="shrink-0 px-6 py-5">
          <Dialog.Title className="text-lg">{name}: run history</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="max-h-none min-h-0 flex-1">
          {query.isPending ? <Text color="muted">Loading runs...</Text> : null}
          {query.isError ? (
            <Box>
              <Text role="alert">Could not load runs.</Text>
              <Button
                color="tertiary"
                onClick={() => {
                  void query.refetch();
                }}
                variant="naked"
              >
                Try again
              </Button>
            </Box>
          ) : null}
          {!query.isPending && !query.isError && !query.data.length ? (
            <Text color="muted">
              No runs yet. The next matching event or scheduled time will appear
              here.
            </Text>
          ) : null}
          <div className="divide-border divide-y">
            {query.data?.map((run) => (
              <Box className="py-4" key={run.id}>
                <Flex align="center" justify="between">
                  <Text className="capitalize" fontWeight="medium">
                    {run.status}
                  </Text>
                  <Text color="muted">
                    {new Date(run.startedAt).toLocaleString()}
                  </Text>
                </Flex>
                {run.error ? (
                  <Text className="mt-2" color="danger">
                    {run.error}
                  </Text>
                ) : null}
              </Box>
            ))}
          </div>
        </Dialog.Body>
        <Dialog.Footer className="shrink-0 justify-end gap-3 py-4">
          <Button color="tertiary" onClick={onClose} variant="outline">
            Close
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
