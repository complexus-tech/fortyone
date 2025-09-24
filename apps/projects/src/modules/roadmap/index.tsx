"use client";

import { useState } from "react";
import { BreadCrumbs, Flex, Button, Box, Text } from "ui";
import { RoadmapIcon, PlusIcon, CopyIcon } from "icons";
import { toast } from "sonner";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import {
  useLocalStorage,
  useTerminology,
  useUserRole,
  useCopyToClipboard,
} from "@/hooks";
import { RoadmapGanttBoard } from "@/components/ui/roadmap-gantt-board";
import { ListObjectives } from "@/modules/objectives/components/list-objectives";
import { NewObjectiveDialog } from "@/components/ui";
import { RoadmapLayoutSwitcher } from "@/components/ui/roadmap-layout-switcher";
import type { RoadmapLayoutType } from "./types";

export const RoadmapPage = () => {
  const [_, copyText] = useCopyToClipboard();
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const [layout, setLayout] = useLocalStorage<RoadmapLayoutType>(
    "roadmapLayout",
    "gantt",
  );
  const { data: objectives = [] } = useObjectives();
  const [isOpen, setIsOpen] = useState(false);
  const [isCopied, setIsCopied] = useState(false);

  const renderContent = () => {
    switch (layout) {
      case "gantt":
        return <RoadmapGanttBoard className="h-full" objectives={objectives} />;
      case "list":
        return <ListObjectives objectives={objectives} />;
      default:
        return null;
    }
  };

  return (
    <>
      <HeaderContainer className="justify-between">
        <Flex gap={2}>
          <MobileMenuButton />
          <BreadCrumbs
            breadCrumbs={[
              {
                name: "Roadmap",
                icon: (
                  <RoadmapIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
                ),
              },
            ]}
          />
        </Flex>
        <Flex align="center" gap={1}>
          <RoadmapLayoutSwitcher
            layout={layout}
            setLayout={setLayout}
            className="hidden md:flex"
          />
          <Box className="hidden md:block">
            <Button
              className="mr-1.5 gap-1 px-3"
              color="tertiary"
              leftIcon={<CopyIcon className="h-4" />}
              onClick={async () => {
                await copyText(window.location.href);
                setIsCopied(true);
                toast.info("Success", {
                  description: "Roadmap link copied to clipboard",
                });
                setTimeout(() => {
                  setIsCopied(false);
                }, 5000);
              }}
              size="sm"
            >
              <span className="hidden md:inline">
                {isCopied ? "Copied" : "Copy link"}
              </span>
              <span className="md:hidden">{isCopied ? "Copied" : "Copy"}</span>
            </Button>
          </Box>
          <Button
            color="invert"
            disabled={userRole === "guest"}
            leftIcon={
              <PlusIcon className="h-[1.1rem] text-current dark:text-current" />
            }
            onClick={() => {
              if (userRole !== "guest") {
                setIsOpen(true);
              }
            }}
            size="sm"
          >
            New {getTermDisplay("objectiveTerm", { capitalize: true })}
          </Button>
        </Flex>
      </HeaderContainer>

      <Box className="h-[calc(100dvh-4rem)]">
        {objectives.length === 0 ? (
          <Box className="flex h-full items-center justify-center">
            <Box className="flex flex-col items-center">
              <RoadmapIcon className="h-12 w-auto" strokeWidth={1.3} />
              <Text className="mb-6 mt-8" fontSize="3xl">
                Your strategic Roadmap awaits
              </Text>
              <Text className="mb-6 max-w-md text-center" color="muted">
                Create {getTermDisplay("objectiveTerm", { variant: "plural" })}{" "}
                to visualize your team&apos;s strategic work and track progress
                toward your goals.
              </Text>
              <Flex gap={2}>
                <Button
                  color="tertiary"
                  disabled={userRole === "guest"}
                  leftIcon={<PlusIcon className="h-[1.1rem]" />}
                  onClick={() => {
                    if (userRole !== "guest") {
                      setIsOpen(true);
                    }
                  }}
                  size="md"
                >
                  Set your first {getTermDisplay("objectiveTerm")}
                </Button>
              </Flex>
            </Box>
          </Box>
        ) : (
          renderContent()
        )}
      </Box>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
