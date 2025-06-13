import React, { useEffect, useMemo } from "react";
import { Button, Dialog, Flex, Tooltip } from "ui";
import {
  ArrowDown2Icon,
  ArrowLeft2Icon,
  ArrowUp2Icon,
  CloseIcon,
  MaximizeIcon,
} from "icons";
import { useStoryById } from "@/modules/story/hooks/story";
import { slugify } from "@/utils";
import type { Story } from "@/modules/stories/types";
// eslint-disable-next-line import/no-cycle -- this is a circular dependency will be fixed in the future
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

  // Memoized navigation state
  const navigationState = useMemo(() => {
    const currentIndex = stories.findIndex((s) => s.id === storyId);
    const hasPrev = currentIndex > 0;
    const hasNext = currentIndex < stories.length - 1;
    return { currentIndex, hasPrev, hasNext };
  }, [stories, storyId]);

  const { currentIndex, hasPrev, hasNext } = navigationState;

  const handlePrev = () => {
    if (hasPrev && onNavigate) {
      onNavigate(stories[currentIndex - 1].id);
    }
  };

  const handleNext = () => {
    if (hasNext && onNavigate) {
      onNavigate(stories[currentIndex + 1].id);
    }
  };

  // Keyboard navigation
  useEffect(() => {
    if (!isOpen || !onNavigate) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      // Only handle navigation if no input elements are focused
      const activeElement = document.activeElement;
      const isInputFocused =
        activeElement?.tagName === "INPUT" ||
        activeElement?.tagName === "TEXTAREA" ||
        (activeElement as HTMLElement).contentEditable === "true";

      if (isInputFocused) return;

      if (event.key === "ArrowUp" || event.key === "ArrowLeft") {
        event.preventDefault();
        handlePrev();
      } else if (event.key === "ArrowDown" || event.key === "ArrowRight") {
        event.preventDefault();
        handleNext();
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen, onNavigate, hasPrev, hasNext, handlePrev, handleNext]);

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content
        className="dark:bg-dark-300 md:mb-auto md:mt-auto"
        hideClose
        size="xl"
      >
        <Dialog.Title className="hidden">
          <span className="sr-only">{story?.title}</span>
        </Dialog.Title>
        <Dialog.Body className="h-[85.5dvh] max-h-[85.5dvh] overflow-y-hidden px-0 pt-0">
          <StoryPage
            isDialog
            isNotifications={false}
            mainHeader={
              <Flex
                className="sticky top-0 z-[2] bg-white/80 px-8 py-4 backdrop-blur dark:bg-dark-300/80"
                gap={2}
                justify="between"
              >
                <Flex align="center" gap={2}>
                  <Button
                    color="tertiary"
                    leftIcon={<ArrowLeft2Icon className="h-[1.1rem]" />}
                    onClick={() => {
                      setIsOpen(false);
                    }}
                    size="sm"
                    variant="naked"
                  >
                    Go back
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
                      size="sm"
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
                      size="sm"
                      variant="naked"
                    >
                      <span className="sr-only">Next</span>
                    </Button>
                  </Tooltip>
                </Flex>
                <Flex align="center" gap={2}>
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
