"use client";

import { useEffect, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { Badge, Box, Button, Flex, Text } from "ui";
import { ArrowRight2Icon } from "icons";
import { toast } from "sonner";
import { useTerminology, useWorkspacePath } from "@/hooks";
import {
  useCreateFigmaInstallSession,
  useDisconnectFigma,
  useFigmaIntegration,
} from "@/lib/hooks/figma";
import {
  SectionHeader,
  SettingsBackButton,
} from "@/modules/settings/components";
import { FigmaIcon } from "./icon";

export const FigmaIntegrationSettings = () => {
  const searchParams = useSearchParams();
  const { data: integration } = useFigmaIntegration();
  const connect = useCreateFigmaInstallSession();
  const disconnect = useDisconnectFigma();
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const handledCallbackResult = useRef<string | null>(null);

  useEffect(() => {
    const connected = searchParams.get("connected") === "1";
    const connectionError = searchParams.get("figma_error");
    if (!connected && !connectionError) return;

    const callbackResult = connected ? "connected" : connectionError;
    if (handledCallbackResult.current === callbackResult) return;
    handledCallbackResult.current = callbackResult;
    if (connected) {
      toast.success("Figma connected");
    } else {
      toast.error("Figma could not be connected. Please try again.");
    }
    const url = new URL(window.location.href);
    url.searchParams.delete("connected");
    url.searchParams.delete("figma_error");
    window.history.replaceState({}, "", url.toString());
  }, [searchParams]);

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
    </Box>
  );
};
