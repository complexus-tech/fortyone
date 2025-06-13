import React from "react";
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
import { StoryPage } from "../../../modules/story";

export const StoryDialog = ({
  isOpen,
  setIsOpen,
  storyId,
}: {
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
  storyId: string;
}) => {
  const { data: story } = useStoryById(storyId);
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

                  <Tooltip side="bottom" title="Previous">
                    <Button
                      asIcon
                      className="ml-5"
                      color="tertiary"
                      leftIcon={<ArrowUp2Icon />}
                      size="sm"
                      variant="naked"
                    >
                      <span className="sr-only">Previous</span>
                    </Button>
                  </Tooltip>
                  <Tooltip side="bottom" title="Next">
                    <Button
                      asIcon
                      color="tertiary"
                      leftIcon={<ArrowDown2Icon />}
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
