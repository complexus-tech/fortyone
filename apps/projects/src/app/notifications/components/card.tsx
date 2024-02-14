import { Avatar, Flex, Text } from "ui";
import { RowWrapper } from "@/components/ui";

export const Card = () => {
  return (
    <RowWrapper className="group block cursor-pointer px-5 transition dark:bg-dark-100/[0.15] dark:hover:bg-dark-100/40 focus:dark:bg-dark-100/40">
      <Flex align="center" className="mb-3" gap={1} justify="between">
        <Flex align="center" gap={1}>
          <Avatar name="John Doe" size="sm" />
          <Text color="muted" textOverflow="truncate">
            <span className="font-medium">Joseph Mukorivo</span> updated an
            issue
          </Text>
        </Flex>
        <Text color="muted">08:59</Text>
      </Flex>
      <Text color="muted" textOverflow="truncate">
        This is from the descition of the issue. This text can very long so be
        careful.
      </Text>
    </RowWrapper>
  );
};
