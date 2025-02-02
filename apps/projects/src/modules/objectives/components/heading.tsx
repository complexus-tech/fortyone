import { Flex, Text, Box, Container } from "ui";

export const TableHeader = ({ isInTeam }: { isInTeam?: boolean }) => {
  return (
    <Box className="sticky top-0 z-[1] border-b-[0.5px] border-gray-100/60 bg-gray-50/60 py-3 backdrop-blur dark:border-dark-100 dark:bg-dark-300/90">
      <Container className="flex items-center justify-between">
        <Box className="w-[300px] shrink-0">
          <Text color="muted" fontWeight="medium">
            Name
          </Text>
        </Box>
        <Flex gap={4}>
          {!isInTeam && (
            <Text
              className="w-[45px] shrink-0"
              color="muted"
              fontWeight="medium"
            >
              Team
            </Text>
          )}

          <Text className="w-[40px] shrink-0" color="muted" fontWeight="medium">
            Lead
          </Text>

          <Text className="w-[60px] shrink-0" color="muted" fontWeight="medium">
            Progress
          </Text>
          <Text
            className="w-[90px] shrink-0 pl-2"
            color="muted"
            fontWeight="medium"
          >
            Status
          </Text>
          <Text
            className="w-[85px] shrink-0 pl-2.5"
            color="muted"
            fontWeight="medium"
          >
            Priority
          </Text>

          <Text
            className="w-[100px] shrink-0 pl-2"
            color="muted"
            fontWeight="medium"
          >
            Target
          </Text>

          <Box className="w-[100px] shrink-0">
            <Text color="muted" fontWeight="medium">
              Health
            </Text>
          </Box>
        </Flex>
      </Container>
    </Box>
  );
};
