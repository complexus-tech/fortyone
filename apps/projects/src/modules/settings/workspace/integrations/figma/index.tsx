"use client";

import { useState } from "react";
import { Badge, Box, Button, Dialog, Flex, Input, Text } from "ui";
import { ArrowRight2Icon, LinkIcon, PlusIcon } from "icons";
import { toast } from "sonner";
import { NewStoryDialog } from "@/components/ui/new-story-dialog";
import { useTerminology, useWorkspacePath } from "@/hooks";
import {
  useCreateFigmaInstallSession,
  useDisconnectFigma,
  useFigmaIntegration,
  useLinkFigmaStory,
  useResolveFigmaLink,
} from "@/lib/hooks/figma";
import type { DetailedStory } from "@/modules/story/types";
import {
  SectionHeader,
  SettingsBackButton,
} from "@/modules/settings/components";
import type { FigmaArtifact } from "./types";
import { FigmaIcon } from "./icon";

export const FigmaIntegrationSettings = () => {
  const { data: integration } = useFigmaIntegration();
  const connect = useCreateFigmaInstallSession();
  const disconnect = useDisconnectFigma();
  const resolveLink = useResolveFigmaLink();
  const linkStory = useLinkFigmaStory();
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const [isSourceDialogOpen, setIsSourceDialogOpen] = useState(false);
  const [isStoryDialogOpen, setIsStoryDialogOpen] = useState(false);
  const [url, setURL] = useState("");
  const [artifact, setArtifact] = useState<FigmaArtifact | null>(null);

  const handleResolve = async () => {
    try {
      const resolvedArtifact = await resolveLink.mutateAsync(url);
      setArtifact(resolvedArtifact);
      setIsSourceDialogOpen(false);
      setIsStoryDialogOpen(true);
    } catch (error) {
      toast.error("Could not open that Figma design", {
        description:
          error instanceof Error
            ? error.message
            : "Try a direct file or frame link.",
      });
    }
  };

  const handleStoryCreated = async (story: DetailedStory) => {
    if (!artifact) return;
    await linkStory.mutateAsync({
      storyId: story.id,
      url: artifact.canonicalUrl,
    });
    toast.success(
      `${getTermDisplay("storyTerm", { capitalize: true })} created from Figma`,
    );
    setArtifact(null);
    setURL("");
  };

  return (
    <Box>
      <Flex align="center" className="mb-6" gap={2}>
        <SettingsBackButton
          href={withWorkspace("/settings/workspace/integrations")}
          label="Back to integrations"
        />
        <Text as="h1" className="text-2xl font-medium">
          Figma
        </Text>
      </Flex>

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          action={
            integration?.connection ? (
              <Flex gap={2}>
                <Button
                  color="tertiary"
                  disabled={disconnect.isPending}
                  onClick={() => {
                    disconnect.mutate();
                  }}
                  variant="outline"
                >
                  Disconnect
                </Button>
                <Button
                  color="invert"
                  onClick={() => {
                    connect.mutate();
                  }}
                >
                  Reconnect
                </Button>
              </Flex>
            ) : (
              <Button
                color="invert"
                disabled={!integration?.configured || connect.isPending}
                onClick={() => {
                  connect.mutate();
                }}
              >
                Connect Figma
              </Button>
            )
          }
          description={`Attach design context to ${getTermDisplay("storyTerm", { variant: "plural" })} and keep handoff status in sync.`}
          title="Connection"
        />
        <Flex align="center" className="px-6 py-5" justify="between">
          <Flex align="center" gap={3}>
            <Flex
              align="center"
              className="bg-surface-muted size-10 rounded-xl"
              justify="center"
            >
              <FigmaIcon className="h-6 w-auto" />
            </Flex>
            <Box>
              <Text className="font-medium">
                {integration?.connection?.handle ??
                  "No Figma account connected"}
              </Text>
              <Text color="muted">
                {integration?.connection?.email ??
                  (integration?.configured
                    ? "Connect an account that can access your design files."
                    : "Add the Figma OAuth credentials to enable this integration.")}
              </Text>
            </Box>
          </Flex>
          <Badge color={integration?.connection ? "success" : "secondary"}>
            {integration?.connection ? "Connected" : "Not connected"}
          </Badge>
        </Flex>
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          action={
            <Button
              disabled={!integration?.connection}
              leftIcon={<PlusIcon />}
              onClick={() => {
                setIsSourceDialogOpen(true);
              }}
            >
              Create from Figma
            </Button>
          }
          description={`Start a new ${storyTerm} from a Figma file or frame and add the backlink automatically.`}
          title={`Create ${storyTerm} from a design`}
        />
        <Flex align="center" className="px-6 py-5" gap={3}>
          <LinkIcon className="text-icon h-5" />
          <Text color="muted">
            Paste an exact Figma frame link. You can review the title, team and
            workflow state before creating the {storyTerm}.
          </Text>
        </Flex>
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          description="Linked frames receive a FortyOne backlink and update their handoff status from Figma Dev Mode."
          title="Design handoff"
        />
        <Box className="divide-border divide-y-[0.5px]">
          {[
            "Rich file and frame previews on work items",
            "FortyOne links inside Figma Dev Mode",
            "Ready for development and completed status updates",
            "Design-change activity without moving workflow status",
          ].map((capability) => (
            <Flex align="center" className="px-6 py-4" gap={3} key={capability}>
              <ArrowRight2Icon className="text-icon h-4" />
              <Text>{capability}</Text>
            </Flex>
          ))}
        </Box>
      </Box>

      <Dialog onOpenChange={setIsSourceDialogOpen} open={isSourceDialogOpen}>
        <Dialog.Content>
          <Dialog.Header className="px-6 pt-6">
            <Dialog.Title>Create from Figma</Dialog.Title>
            <Dialog.Description>
              Paste a direct file or frame URL to bring its design context into
              FortyOne.
            </Dialog.Description>
          </Dialog.Header>
          <Dialog.Body className="space-y-4 pb-6">
            <Input
              autoFocus
              label="Figma URL"
              onChange={(event) => {
                setURL(event.target.value);
              }}
              placeholder="https://www.figma.com/design/..."
              type="url"
              value={url}
            />
            <Flex justify="end">
              <Button
                disabled={!url.trim() || resolveLink.isPending}
                onClick={() => void handleResolve()}
              >
                Continue
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>

      <NewStoryDialog
        description={artifact ? `Design: ${artifact.canonicalUrl}` : undefined}
        initialTitle={artifact?.nodeName ?? artifact?.fileName}
        isOpen={isStoryDialogOpen}
        onCreated={handleStoryCreated}
        setIsOpen={setIsStoryDialogOpen}
      />
    </Box>
  );
};
