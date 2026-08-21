"use client";

import { useState } from "react";
import { Box, Button, Flex, Input, Popover, Text } from "ui";
import { CloseIcon } from "icons";
import { toast } from "sonner";
import { useFigmaIntegration, useResolveFigmaLink } from "@/lib/hooks/figma";
import { FigmaIcon } from "@/modules/settings/workspace/integrations/figma/icon";
import type { FigmaArtifact } from "@/modules/settings/workspace/integrations/figma/types";

export const NewStoryFigmaSource = ({
  artifact,
  enabled,
  onArtifactChange,
  onTitleSuggestion,
}: {
  artifact: FigmaArtifact | null;
  enabled: boolean;
  onArtifactChange: (artifact: FigmaArtifact | null) => void;
  onTitleSuggestion: (title: string) => void;
}) => {
  const { data: integration } = useFigmaIntegration({ enabled });
  const resolveLink = useResolveFigmaLink();
  const [isOpen, setIsOpen] = useState(false);
  const [url, setURL] = useState("");

  if (!integration?.connection?.isActive) return null;

  const handleResolve = async () => {
    try {
      const resolvedArtifact = await resolveLink.mutateAsync(url.trim());
      onArtifactChange(resolvedArtifact);
      onTitleSuggestion(resolvedArtifact.nodeName ?? resolvedArtifact.fileName);
      setURL(resolvedArtifact.canonicalUrl);
      setIsOpen(false);
    } catch (error) {
      toast.error("Could not open that Figma design", {
        description:
          error instanceof Error
            ? error.message
            : "Try a direct Figma file or frame link.",
      });
    }
  };

  return (
    <Box className="order-12">
      <Popover modal onOpenChange={setIsOpen} open={isOpen}>
        <Popover.Trigger asChild>
          <Button
            className="dark:bg-surface-elevated/90 max-w-52 gap-1.5 px-2"
            color="tertiary"
            leftIcon={<FigmaIcon className="h-4 w-auto shrink-0" />}
            size="sm"
            type="button"
            variant="outline"
          >
            <span className="truncate">
              {artifact
                ? artifact.nodeName ?? artifact.fileName
                : "Figma design"}
            </span>
          </Button>
        </Popover.Trigger>
        <Popover.Content align="end" className="w-96 p-4">
          <Flex align="start" gap={3} justify="between">
            <Box>
              <Text className="font-medium">Create from Figma</Text>
              <Text color="muted">
                Paste a direct file or frame URL to attach after creation.
              </Text>
            </Box>
            {artifact ? (
              <Button
                asIcon
                color="tertiary"
                onClick={() => {
                  onArtifactChange(null);
                  setURL("");
                  setIsOpen(false);
                }}
                size="xs"
                title="Remove Figma design"
                variant="naked"
              >
                <CloseIcon className="h-4 w-auto" />
              </Button>
            ) : null}
          </Flex>
          <Input
            autoFocus
            className="mt-4"
            label="Figma URL"
            onChange={(event) => {
              setURL(event.target.value);
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter" && url.trim()) {
                event.preventDefault();
                void handleResolve();
              }
            }}
            placeholder="https://www.figma.com/design/..."
            type="url"
            value={url}
          />
          <Flex className="mt-4" justify="end">
            <Button
              color="tertiary"
              disabled={!url.trim() || resolveLink.isPending}
              loading={resolveLink.isPending}
              loadingText="Opening design..."
              onClick={() => void handleResolve()}
              size="sm"
            >
              Use design
            </Button>
          </Flex>
        </Popover.Content>
      </Popover>
    </Box>
  );
};
