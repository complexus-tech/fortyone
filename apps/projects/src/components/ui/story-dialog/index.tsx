import React, { useEffect, useMemo } from "react";
import { Button, Dialog, Flex, Tooltip } from "ui";
import {
  ArrowDown2Icon,
  ArrowLeft2Icon,
  ArrowUp2Icon,
  CloseIcon,
  MaximizeIcon,
} from "icons";
import { useQueryClient } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useHotkeys } from "react-hotkeys-hook";
import { useWorkspacePath } from "@/hooks";
import { useStoryById } from "@/modules/story/hooks/story";
import { storyKeys } from "@/modules/stories/constants";
import { getStory } from "@/modules/story/queries/get-story";
import { slugify } from "@/utils";
import type { Story } from "@/modules/stories/types";
import { StoryPage } from "../../../modules/story";

export const StoryDialog = ({
  isOpen,
  setIsOpen,
  storyId,
  stories = [],
  onNavigate,
}: {
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
  storyId: string;
  stories?: Story[];
  onNavigate?: (storyId: string) => void;
}) => {
  const { data: story } = useStoryById(storyId);
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  // Memoized navigation state
  const navigationState = useMemo(() => {
    const currentIndex = stories.findIndex((s) => s.id === storyId);
    const hasPrev = currentIndex > 0;
    const hasNext = currentIndex < stories.length - 1;
    const prevStoryId = hasPrev ? stories[currentIndex - 1]?.id : null;
    const nextStoryId = hasNext ? stories[currentIndex + 1]?.id : null;
    return { hasPrev, hasNext, prevStoryId, nextStoryId };
  }, [stories, storyId]);

  const { hasPrev, hasNext, prevStoryId, nextStoryId } = navigationState;

  // Prefetch adjacent stories
  useEffect(() => {
    if (!session || !isOpen) return;

    const ctx = { session, workspaceSlug };
    const prefetchStory = async (storyId: string) => {
      // Only prefetch if the story isn't already cached
      const cachedStory = queryClient.getQueryData(
        storyKeys.detail(workspaceSlug, storyId),
      );
      if (!cachedStory) {
        await queryClient.prefetchQuery({
          queryKey: storyKeys.detail(workspaceSlug, storyId),
          queryFn: () => getStory(storyId, ctx),
          staleTime: 3 * 60 * 1000, // 3 minutes
        });
      }
    };

    // Prefetch previous story
    if (prevStoryId) {
      prefetchStory(prevStoryId);
    }

    // Prefetch next story
    if (nextStoryId) {
      prefetchStory(nextStoryId);
    }
  }, [queryClient, session, workspaceSlug, isOpen, prevStoryId, nextStoryId]);

  const handlePrev = () => {
    if (hasPrev && onNavigate && prevStoryId) {
      onNavigate(prevStoryId);
    }
  };

  const handleNext = () => {
    if (hasNext && onNavigate && nextStoryId) {
      onNavigate(nextStoryId);
    }
  };

  // Keyboard navigation using react-hotkeys-hook
  useHotkeys(
    "up, left",
    () => {
      if (isOpen && onNavigate) {
        handlePrev();
      }
    },
    [isOpen, onNavigate, handlePrev],
  );

  useHotkeys(
    "down, right",
    () => {
      if (isOpen && onNavigate) {
        handleNext();
      }
    },
    [isOpen, onNavigate, handleNext],
  );

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content
        className="border-border bg-surface max-w-[88rem] overflow-clip rounded-[1.25rem] border md:mt-auto md:mb-auto"
        hideClose
        overlayClassName="bg-black/10"
      >
        <Dialog.Title className="hidden">
          <span className="sr-only">{story?.title}</span>
        </Dialog.Title>
        <Dialog.Body className="h-[86dvh] max-h-[86dvh] overflow-y-hidden px-0 pt-0">
          <StoryPage
            isDialog
            isNotifications={false}
            mainHeader={
              <Flex
                className="bg-surface/80 sticky top-0 z-2 px-6 py-4 backdrop-blur"
                gap={2}
                justify="between"
              >
                <Flex align="center" gap={2}>
                  <Button
                    className="text-foreground/80 border-border bg-surface-muted/30 shrink-0 gap-1 pr-4 pl-2.5 font-semibold tracking-wide"
                    color="tertiary"
                    leftIcon={<ArrowLeft2Icon strokeWidth={2.9} />}
                    onClick={() => {
                      setIsOpen(false);
                    }}
                    variant="naked"
                  >
                    Close
                  </Button>

                  <Tooltip
                    side="bottom"
                    title={hasPrev ? "Previous (↑ or ←)" : "Previous"}
                  >
                    <Button
                      asIcon
                      className="ml-5"
                      color="tertiary"
                      disabled={!hasPrev || !onNavigate}
                      leftIcon={<ArrowUp2Icon />}
                      onClick={handlePrev}
                      variant="naked"
                    >
                      <span className="sr-only">Previous</span>
                    </Button>
                  </Tooltip>
                  <Tooltip
                    side="bottom"
                    title={hasNext ? "Next (↓ or →)" : "Next"}
                  >
                    <Button
                      asIcon
                      color="tertiary"
                      disabled={!hasNext || !onNavigate}
                      leftIcon={<ArrowDown2Icon />}
                      onClick={handleNext}
                      variant="naked"
                    >
                      <span className="sr-only">Next</span>
                    </Button>
                  </Tooltip>
                </Flex>
                <Flex align="center" className="hidden" gap={2}>
                  <Tooltip side="bottom" title="Fullscreen">
                    <span>
                      <Button
                        asIcon
                        color="tertiary"
                        href={`/story/${story?.id}/${slugify(story?.title)}`}
                        leftIcon={
                          <MaximizeIcon
                            className="h-[1.15rem]"
                            strokeWidth={2.5}
                          />
                        }
                        size="sm"
                        variant="naked"
                      >
                        <span className="sr-only">Fullscreen</span>
                      </Button>
                    </span>
                  </Tooltip>
                  <Tooltip side="bottom" title="Close">
                    <Button
                      asIcon
                      color="tertiary"
                      leftIcon={<CloseIcon strokeWidth={2.8} />}
                      onClick={() => {
                        setIsOpen(false);
                      }}
                      size="sm"
                      variant="naked"
                    >
                      <span className="sr-only">Close</span>
                    </Button>
                  </Tooltip>
                </Flex>
              </Flex>
            }
            storyId={storyId}
          />
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
