import { Avatar, Box, Flex, Text } from "ui";
import { cn } from "lib";
import { FORMER_USER_NAME } from "@/lib/former-user";

// Intentionally has no profile link, tooltip, or external avatar source.
export const FormerUser = ({
  avatarSurfaceClassName,
}: {
  avatarSurfaceClassName?: string;
}) => (
  <Flex align="center" className="shrink-0" gap={1}>
    <Box
      className={cn(
        "bg-surface flex aspect-square items-center rounded-full p-[0.3rem]",
        avatarSurfaceClassName,
      )}
    >
      <Avatar name={FORMER_USER_NAME} size="xs" />
    </Box>
    <Text className="ml-1 text-sm md:text-[0.95rem]" fontWeight="medium">
      {FORMER_USER_NAME}
    </Text>
  </Flex>
);
