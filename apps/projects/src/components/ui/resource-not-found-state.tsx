import type { ReactNode } from "react";
import { ArrowLeft2Icon } from "icons";
import { Box, Button, Text } from "ui";
import { NotFoundIllustration } from "./illustrations/empty-state-illustrations";

export const ResourceNotFoundState = ({
  actionLabel = "Go to my work",
  description,
  href,
  illustration,
  title,
}: {
  actionLabel?: string;
  description: ReactNode;
  href: string;
  illustration?: ReactNode;
  title: ReactNode;
}) => (
  <Box className="flex h-screen items-center justify-center px-6">
    <Box className="flex flex-col items-center">
      {illustration ?? <NotFoundIllustration />}
      <Text className="mt-8 mb-4 text-center" fontSize="3xl">
        {title}
      </Text>
      <Text className="mb-6 max-w-md text-center" color="muted">
        {description}
      </Text>
      <Button
        className="gap-1 pl-2"
        color="tertiary"
        href={href}
        leftIcon={<ArrowLeft2Icon className="h-[1.05rem] w-auto" />}
      >
        {actionLabel}
      </Button>
    </Box>
  </Box>
);
