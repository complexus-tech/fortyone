"use client";

import { useState } from "react";
import { Box, Button, Flex, Input, Menu, Popover, Text } from "ui";
import { AiIcon, CloseIcon, MoreHorizontalIcon } from "icons";
import { toast } from "sonner";
import {
  useExtractFigmaDescription,
  useFigmaIntegration,
  useResolveFigmaLink,
} from "@/lib/hooks/figma";
import { FigmaIcon } from "@/modules/settings/workspace/integrations/figma/icon";
import type { FigmaArtifact } from "@/modules/settings/workspace/integrations/figma/types";
import type { FigmaDescription } from "@/modules/settings/workspace/integrations/figma/description";

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
  onAddDescription: (description: FigmaDescription) => void;
  onArtifactsChange: (artifacts: FigmaArtifact[]) => void;
  onTitleSuggestion: (title: string) => void;
}) => {
  const { data: integration } = useFigmaIntegration({ enabled });
  const resolveLink = useResolveFigmaLink();
  const extractDescription = useExtractFigmaDescription();
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

  const handleExtractDescription = async (artifact: FigmaArtifact) => {
    const toastId = toast.loading("Extracting story description...");
    try {
      const description = await extractDescription.mutateAsync(artifact);
      onAddDescription(description);
      toast.success("Description added", {
        description: "Review the AI-generated draft before creating the story.",
        id: toastId,
      });
    } catch (error) {
      toast.error("Description could not be extracted", {
        description:
          error instanceof Error ? error.message : "Please try again.",
        id: toastId,
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
          className="w-[32rem] max-w-[calc(100vw-2rem)] p-5"
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
                    className="border-border bg-surface-muted/50 overflow-hidden rounded-lg border py-3 pr-3 pl-4"
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
                    </Box>
                    <Menu>
                      <Menu.Button>
                        <button
                          aria-label={`More actions for ${artifactName(artifact)}`}
                          className="text-icon hover:bg-surface-muted rounded-md p-1.5"
                          type="button"
                        >
                          <MoreHorizontalIcon className="h-5 w-auto" />
                        </button>
                      </Menu.Button>
                      <Menu.Items align="end" className="w-52">
                        <Menu.Group>
                          <Menu.Item
                            disabled={
                              !artifact.textContent?.length ||
                              extractDescription.isPending
                            }
                            onSelect={() => {
                              void handleExtractDescription(artifact);
                            }}
                          >
                            <AiIcon />
                            Extract description
                          </Menu.Item>
                          <Menu.Item
                            onSelect={() => {
                              onArtifactsChange(
                                artifacts.filter(
                                  ({ canonicalUrl }) =>
                                    canonicalUrl !== artifact.canonicalUrl,
                                ),
                              );
                            }}
                          >
                            <CloseIcon />
                            Remove design
                          </Menu.Item>
                        </Menu.Group>
                      </Menu.Items>
                    </Menu>
                  </Flex>
                );
              })}
            </Box>
          ) : null}

          <Box className="mt-4">
            <Input
              autoFocus
              className="dark:bg-surface/60"
              label={
                artifacts.length > 0 ? "Add another Figma URL" : "Figma URL"
              }
              labelClassName="mb-1"
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
          <Flex className="mt-4" justify="end">
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
