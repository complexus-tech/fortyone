"use client";
import { Box, Button, Text } from "ui";
import { cn } from "lib";
import type { ReactNode } from "react";
import dynamic from "next/dynamic";
import { useWorkspacePath } from "@/hooks";
import { NotFoundIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { ResourceNotFoundState } from "@/components/ui/resource-not-found-state";
import { MainDetailsSkeleton } from "./components/main-details-skeleton";
import { Options } from "./components/options";
import { useStoryById } from "./hooks/story";
import { StorySkeleton } from "./components/story-skeleton";

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
  const {
    data: story,
    isError,
    isFetching,
    isPending,
    refetch,
  } = useStoryById(storyId);
  const { withWorkspace } = useWorkspacePath();

  if (isPending) {
    return <StorySkeleton isDialog={isDialog} />;
  }

  if (isError) {
    return (
      <Box className="flex h-screen items-center justify-center">
        <Box className="flex flex-col items-center">
          <NotFoundIllustration />
          <Text className="mt-10 mb-6" fontSize="3xl">
            Unable to load this item
          </Text>
          <Text className="mb-6 max-w-md text-center" color="muted">
            Something went wrong while loading this item. Try again in a moment.
          </Text>
          <Button
            color="tertiary"
            loading={isFetching}
            onClick={() => void refetch()}
          >
            Try again
          </Button>
        </Box>
      </Box>
    );
  }

  return (
    <Box
      className={cn("h-dvh", {
        "h-[85dvh] overflow-y-auto": isDialog,
      })}
    >
      {story ? (
        <>
          <Box
            className={cn("md:hidden", {
              "dark:bg-surface": isDialog,
            })}
          >
            <MainDetails
              isDialog={isDialog}
              isNotifications={Boolean(isNotifications)}
              storyId={storyId}
            />
          </Box>
          <Box
            className={cn("hidden h-full md:flex", {
              "notification-story-container": isNotifications,
            })}
          >
            <Box
              className={cn("min-w-0 flex-1", {
                "dark:bg-surface": isDialog,
              })}
            >
              <MainDetails
                isDialog={isDialog}
                isNotifications={Boolean(isNotifications)}
                mainHeader={mainHeader}
                storyId={storyId}
              />
            </Box>
            <Box
              className={cn(
                "border-border w-(--story-sidebar-width) shrink-0 border-l-[0.5px]",
                {
                  "notification-story-sidebar": isNotifications,
                },
              )}
            >
              <Options
                isDialog={isDialog}
                isNotifications={Boolean(isNotifications)}
                storyId={storyId}
              />
            </Box>
          </Box>
        </>
      ) : (
        <ResourceNotFoundState
          description="This item might not exist or you do not have access to it."
          href={withWorkspace("/my-work")}
          title="404: Item not found"
        />
      )}
    </Box>
  );
};
