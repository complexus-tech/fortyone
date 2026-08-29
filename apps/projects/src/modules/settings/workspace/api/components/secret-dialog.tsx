"use client";

import { CheckIcon, CopyIcon } from "icons";
import { Box, Button, Dialog, Flex, Text } from "ui";
import { useCopyToClipboard } from "@/hooks";

type SecretDialogProps = {
  description: string;
  label: string;
  onClose: () => void;
  open: boolean;
  secret: string;
  title: string;
  values?: { label: string; value: string }[];
};

export const SecretDialog = ({
  description,
  label,
  onClose,
  open,
  secret,
  title,
  values = [],
}: SecretDialogProps) => {
  const [copied, copy] = useCopyToClipboard();

  return (
    <Dialog
      onOpenChange={(nextOpen) => {
        if (!nextOpen) onClose();
      }}
      open={open}
    >
      <Dialog.Content className="max-w-2xl" size="md">
        <Dialog.Header className="px-6 pt-6 pb-2">
          <Dialog.Title className="text-xl">{title}</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-4 px-6 pt-2 pb-4">
          <Box className="border-warning/30 bg-warning/5 rounded-xl border p-4">
            <Text className="font-medium">Shown once</Text>
            <Text color="muted">{description}</Text>
          </Box>

          {values.map((value) => (
            <Box key={value.label}>
              <Text className="mb-1.5" color="muted">
                {value.label}
              </Text>
              <Flex
                align="center"
                className="border-border bg-surface-elevated rounded-xl border px-3 py-2"
                gap={2}
                justify="between"
              >
                <Text className="min-w-0 truncate font-mono">
                  {value.value}
                </Text>
                <Button
                  aria-label={`Copy ${value.label}`}
                  color="tertiary"
                  onClick={() => void copy(value.value)}
                  size="sm"
                  variant="naked"
                >
                  <CopyIcon className="h-4" />
                </Button>
              </Flex>
            </Box>
          ))}

          <Box>
            <Text className="mb-1.5" color="muted">
              {label}
            </Text>
            <Box className="border-border bg-surface-elevated rounded-xl border p-3">
              <Text className="font-mono break-all">{secret}</Text>
            </Box>
            <Button
              className="mt-2"
              color="tertiary"
              leftIcon={
                copied === secret ? (
                  <CheckIcon className="h-4" />
                ) : (
                  <CopyIcon className="h-4" />
                )
              }
              onClick={() => void copy(secret)}
              size="sm"
              variant="outline"
            >
              {copied === secret ? "Copied" : `Copy ${label.toLowerCase()}`}
            </Button>
          </Box>
        </Dialog.Body>
        <Dialog.Footer className="justify-end">
          <Button onClick={onClose}>I have stored it</Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
