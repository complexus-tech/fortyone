"use client";
import { Avatar, Flex, Text } from "ui";
import { cn } from "lib";
import { RowWrapper } from "@/components/ui";

export const NotificationCard = ({ read }: { read: boolean }) => {
  return (
    <RowWrapper
      className={cn(
        "group block cursor-pointer border-l-2 px-5 transition dark:border-dark-100/70 dark:bg-dark-100/[0.15] dark:hover:bg-dark-100/40 focus:dark:bg-dark-100/40",
        {
          "border-l-2 border-l-primary/90 dark:border-l-primary/90": !read,
        },
      )}
    >
      <Flex align="center" className="mb-3" gap={1} justify="between">
        <Flex align="center" className="flex-1 gap-1.5">
          <Text
            className="w-max shrink-0 opacity-80"
            color="muted"
            fontWeight="medium"
            textOverflow="truncate"
          >
            COM-123
          </Text>
          <Text className="opacity-80" textOverflow="truncate">
            My cool title. This is a
          </Text>
        </Flex>
        <Text className="shrink-0" color="muted">
          08:59
        </Text>
      </Flex>

      <Flex align="end" gap={2}>
        <Avatar
          className="aspect-square shrink-0"
          name="John Doe"
          size="sm"
          src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
        />
        <Text className="opacity-80" color="muted" textOverflow="truncate">
          This is from the descition of the issue. This text can very long so be
          careful.
        </Text>
      </Flex>
    </RowWrapper>
  );
};
