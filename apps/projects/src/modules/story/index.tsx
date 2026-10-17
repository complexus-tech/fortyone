"use client";
import { Box, Button, Text } from "ui";
import { cn } from "lib";
import type { ReactNode } from "react";
import { useEffect } from "react";
import dynamic from "next/dynamic";
import { useWorkspacePath } from "@/hooks";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useRecordStoryVisit } from "@/shared/story/use-recent-story-history";
import { NotFoundIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { ResourceNotFoundState } from "@/components/ui/resource-not-found-state";
import { MainDetailsSkeleton } from "./components/main-details-skeleton";
import { Options } from "./components/options";
import { useStoryById } from "./hooks/story";
import { StorySkeleton } from "./components/story-skeleton";
import { loadMainDetails, preloadMainDetails } from "./load-main-details";
import { StoryCloseButton } from "./components/story-close-button";

const StoryState = ({
  children,
  onClose,
  term,
}: {
  children: ReactNode;
  onClose?: () => void;
  term: string;
}) =>
  onClose ? (
    <Box className="relative h-full min-h-0">
      <div className="absolute top-3 right-5 z-10">
        <StoryCloseButton onClose={onClose} term={term} />
      </div>
      {children}
    </Box>
  ) : (
    children
  );

const MainDetails = dynamic(loadMainDetails, {
  ssr: false,
  loading: () => <MainDetailsSkeleton />,
});

export const StoryPage = ({
  storyId,
  isNotifications,
  isDialog,
  mainHeader,
  onClose,
}: {
  storyId: string;
  isNotifications?: boolean;
  isDialog?: boolean;
  mainHeader?: ReactNode;
  onClose?: () => void;
}) => {
  const {
    data: story,
    isError,
    isFetching,
    isPending,
    refetch,
  } = useStoryById(storyId);
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology(story?.teamId ?? null);
  const term = getTermDisplay("storyTerm");
  useRecordStoryVisit(
    !isError && story && !story.deletedAt ? story.id : undefined,
  );

  useEffect(() => {
    // Direct links and immediate clicks should load code alongside ticket data.
    // Importing the component does not create editors before access is resolved.
    void preloadMainDetails();
  }, []);

  if (isPending) {
    return (
      <StoryState onClose={onClose} term={term}>
        <StorySkeleton isDialog={isDialog} />
      </StoryState>
    );
  }

  if (isError) {
    return (
      <StoryState onClose={onClose} term={term}>
        <Box className="flex h-full min-h-0 items-center justify-center">
          <Box className="flex flex-col items-center">
            <NotFoundIllustration />
            <Text className="mt-10 mb-6" fontSize="3xl">
              Unable to load this item
            </Text>
            <Text className="mb-6 max-w-md text-center" color="muted">
              Something went wrong while loading this item. Try again in a
              moment.
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
      </StoryState>
    );
  }

  return (
    <Box
      className={cn("h-full min-h-0", {
        "overflow-y-auto": isDialog,
      })}
    >
      {story ? (
        <Box
          className={cn("h-full min-h-0 md:flex", {
            "md:[container:notification-story/inline-size]": isNotifications,
          })}
        >
          <Box
            className={cn("h-full min-h-0 min-w-0 flex-1", {
              "dark:bg-surface": isDialog,
            })}
          >
            <MainDetails
              isDialog={isDialog}
              isNotifications={Boolean(isNotifications)}
              mainHeader={
                mainHeader ? (
                  <Box className="hidden md:contents">{mainHeader}</Box>
                ) : undefined
              }
              onClose={onClose}
              storyId={storyId}
            />
          </Box>
          <Box
            className={cn(
              "border-border hidden h-full min-h-0 w-(--story-sidebar-width) shrink-0 overflow-hidden border-l-[0.5px] md:block",
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
      ) : (
        <StoryState onClose={onClose} term={term}>
          <ResourceNotFoundState
            description="This item might not exist or you do not have access to it."
            href={withWorkspace("/my-work")}
            title="404: Item not found"
          />
        </StoryState>
      )}
    </Box>
  );
};
