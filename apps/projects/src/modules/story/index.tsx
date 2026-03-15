"use client";
import { Box, Button, Text } from "ui";
import { cn } from "lib";
import type { ReactNode } from "react";
import dynamic from "next/dynamic";
import { ArrowLeft2Icon, StoryMissingIcon } from "icons";
import { MainDetailsSkeleton } from "./components/main-details-skeleton";
import { Options } from "./components/options";
import { useStoryById } from "./hooks/story";
import { StorySkeleton } from "./components/story-skeleton";
import { useWorkspacePath } from "@/hooks";

const MainDetails = dynamic(
  () => import("./components/main-details").then((mod) => mod.MainDetails),
  {
    ssr: false,
    loading: () => <MainDetailsSkeleton />,
  },
);

export const StoryPage = ({
  storyId,
  isNotifications,
  isDialog,
  mainHeader,
}: {
  storyId: string;
  isNotifications?: boolean;
  isDialog?: boolean;
  mainHeader?: ReactNode;
}) => {
  const { isPending, data: story } = useStoryById(storyId);
  const { withWorkspace } = useWorkspacePath();

  if (isPending) {
    return <StorySkeleton isNotifications={isNotifications} />;
  }

  return (
    <Box
      className={cn("h-dvh", {
        "h-[85dvh] overflow-y-auto": isDialog,
      })}
    >
      {story ? (
        <>
          <Box className="md:hidden">
            <MainDetails
              isNotifications={Boolean(isNotifications)}
              storyId={storyId}
            />
          </Box>
          <Box className="hidden h-full md:flex">
            <Box className="min-w-0 flex-1">
                <MainDetails
                  isDialog={isDialog}
                  isNotifications={Boolean(isNotifications)}
                  mainHeader={mainHeader}
                  storyId={storyId}
                />
            </Box>
            <Box className="border-border w-(--story-sidebar-width) shrink-0 border-l-[0.5px]">
                <Options
                  isDialog={isDialog}
                  isNotifications={Boolean(isNotifications)}
                  storyId={storyId}
                />
            </Box>
          </Box>
        </>
      ) : (
        <Box className="flex h-screen items-center justify-center">
          <Box className="flex flex-col items-center">
            <StoryMissingIcon className="h-20 w-auto rotate-12" />
            <Text className="mt-10 mb-6" fontSize="3xl">
              404: Item not found
            </Text>
            <Text className="mb-6 max-w-md text-center" color="muted">
              This item might not exist or you do not have access to it.
            </Text>
            <Button
              className="gap-1 pl-2"
              color="tertiary"
              href={withWorkspace("/my-work")}
              leftIcon={<ArrowLeft2Icon className="h-[1.05rem] w-auto" />}
            >
              Go to my work
            </Button>
          </Box>
        </Box>
      )}
    </Box>
  );
};
