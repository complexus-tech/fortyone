import { cn } from "lib";
import { Box, Text, Dialog, Button, Flex, BlurImage } from "ui";
import { CloseIcon } from "icons";
import { useState } from "react";

export const AttachmentPreview = ({
  name,
  url,
  className,
}: {
  name: string;
  url: string;
  className?: string;
}) => {
  const [isOpen, setOpen] = useState(false);

  return (
    <>
      <Box
        tabIndex={0}
        onClick={() => {
          setOpen(true);
        }}
        className="relative cursor-pointer overflow-hidden"
      >
        <BlurImage
          src={url}
          alt={name}
          className={cn(
            "aspect-[5/4] h-full w-full object-cover object-top",
            className,
          )}
        />
      </Box>
      <Dialog open={isOpen} onOpenChange={setOpen}>
        <Dialog.Content size="lg" className="relative my-auto p-2" hideClose>
          <Dialog.Header className="sr-only">
            <Dialog.Title className="mb-0 px-6">{name}</Dialog.Title>
          </Dialog.Header>
          <Box className="flex h-[70vh] items-center justify-center overflow-y-auto rounded-lg">
            <BlurImage
              src={url}
              alt={name}
              className="h-full"
              imageClassName="object-contain"
            />
          </Box>
          <Box className="pointer-events-none absolute left-0 right-0 top-0 z-[3] h-20 bg-gradient-to-b from-dark/80 px-6 py-5">
            <Flex
              className="pointer-events-auto"
              align="center"
              justify="between"
            >
              <Text className="text-white" fontSize="sm">
                {name}
              </Text>
              <Button
                asIcon
                leftIcon={<CloseIcon className="h-4" />}
                color="tertiary"
                size="xs"
                rounded="full"
                onClick={() => setOpen(false)}
              >
                <span className="sr-only">Close</span>
              </Button>
            </Flex>
          </Box>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
