"use client";

import { useState } from "react";
import { Box, Button, Flex, Input, Popover, Text } from "ui";
import { CloseIcon } from "icons";
import { toast } from "sonner";
import { useFigmaIntegration, useResolveFigmaLink } from "@/lib/hooks/figma";
import { FigmaIcon } from "@/modules/settings/workspace/integrations/figma/icon";
import type { FigmaArtifact } from "@/modules/settings/workspace/integrations/figma/types";

const artifactName = (artifact: FigmaArtifact) =>
  artifact.nodeName ?? artifact.fileName;

export const NewStoryFigmaSource = ({
  artifacts,
  enabled,
  onAddDescription,
  onArtifactsChange,
  onTitleSuggestion,
}: {
  artifacts: FigmaArtifact[];
  enabled: boolean;
  onAddDescription: (artifact: FigmaArtifact) => void;
  onArtifactsChange: (artifacts: FigmaArtifact[]) => void;
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
      if (
        artifacts.some(
          ({ canonicalUrl }) => canonicalUrl === resolvedArtifact.canonicalUrl,
        )
      ) {
        toast.info("That Figma design is already selected");
        return;
      }
      if (artifacts.length === 0) {
        onTitleSuggestion(artifactName(resolvedArtifact));
      }
      onArtifactsChange([...artifacts, resolvedArtifact]);
      setURL("");
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
              {artifacts.length === 0
                ? "From Figma design"
                : `${artifacts.length} design${artifacts.length === 1 ? "" : "s"} selected`}
            </span>
          </Button>
        </Popover.Trigger>
        <Popover.Content
          align="end"
          className="w-[30rem] max-w-[calc(100vw-2rem)] p-5"
        >
          <Box>
            <Text className="mb-1.5 text-lg font-medium">
              Create from Figma
            </Text>
            <Text color="muted">
              Add one or more direct file or frame URLs. Selected designs will
              be attached after the story is created.
            </Text>
          </Box>

          {artifacts.length > 0 ? (
            <Box className="mt-4 space-y-2">
              {artifacts.map((artifact) => {
                return (
                  <Flex
                    align="center"
                    className="border-border bg-surface-muted/50 overflow-hidden rounded-lg border p-2"
                    gap={3}
                    key={artifact.canonicalUrl}
                  >
                    <Box
                      className="bg-surface-muted h-12 w-16 shrink-0 rounded-md bg-cover bg-center"
                      style={
                        artifact.thumbnailUrl
                          ? {
                              backgroundImage: `url("${artifact.thumbnailUrl.replaceAll('"', "%22")}")`,
                            }
                          : undefined
                      }
                    >
                      {!artifact.thumbnailUrl ? (
                        <Flex
                          align="center"
                          className="h-full"
                          justify="center"
                        >
                          <FigmaIcon className="h-5 w-auto" />
                        </Flex>
                      ) : null}
                    </Box>
                    <Box className="min-w-0 flex-1">
                      <Text className="truncate font-medium">
                        {artifactName(artifact)}
                      </Text>
                      <Text className="truncate text-xs" color="muted">
                        {artifact.nodeName
                          ? artifact.fileName
                          : "Entire Figma file"}
                      </Text>
                      {(artifact.textContent?.length ?? 0) > 0 ? (
                        <Button
                          className="mt-1 h-auto p-0 text-xs"
                          color="tertiary"
                          onClick={() => {
                            onAddDescription(artifact);
                          }}
                          size="xs"
                          type="button"
                          variant="naked"
                        >
                          Add design text to description
                        </Button>
                      ) : null}
                    </Box>
                    <Button
                      asIcon
                      color="tertiary"
                      onClick={() => {
                        onArtifactsChange(
                          artifacts.filter(
                            ({ canonicalUrl }) =>
                              canonicalUrl !== artifact.canonicalUrl,
                          ),
                        );
                      }}
                      size="xs"
                      title={`Remove ${artifactName(artifact)}`}
                      type="button"
                      variant="naked"
                    >
                      <CloseIcon className="h-4 w-auto" />
                    </Button>
                  </Flex>
                );
              })}
            </Box>
          ) : null}

          <Box className="mt-4">
            <Input
              autoFocus
              label={
                artifacts.length > 0 ? "Add another Figma URL" : "Figma URL"
              }
              labelClassName="mb-0.5"
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
          </Box>
          <Flex className="mt-4" justify="between">
            <Text className="self-center text-xs" color="muted">
              {artifacts.length > 0
                ? `${artifacts.length} ready to attach`
                : "Frame links provide the richest preview and handoff sync."}
            </Text>
            <Button
              color="tertiary"
              disabled={!url.trim() || resolveLink.isPending}
              loading={resolveLink.isPending}
              loadingText="Opening design..."
              onClick={() => void handleResolve()}
              size="sm"
            >
              {artifacts.length > 0 ? "Add design" : "Use design"}
            </Button>
          </Flex>
        </Popover.Content>
      </Popover>
    </Box>
  );
};
