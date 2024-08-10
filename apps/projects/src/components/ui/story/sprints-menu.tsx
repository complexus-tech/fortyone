import { Box, Button, Flex, Menu, Text } from "ui";
import { PlusIcon, SprintsIcon } from "icons";
import { useSprints } from "@/lib/hooks/sprints";

export const SprintsMenu = () => {
  const { data: sprints = [] } = useSprints();
  return (
    <Menu>
      <Menu.Button>
        <Button
          color="tertiary"
          leftIcon={<PlusIcon className="h-5 w-auto" />}
          size="md"
          variant="naked"
        >
          Add to sprint
        </Button>
      </Menu.Button>
      <Menu.Items align="end" className="w-64">
        <Menu.Group className="px-4">
          <Menu.Input autoFocus placeholder="Select sprint..." />
        </Menu.Group>
        <Menu.Separator className="my-2" />
        <Menu.Group>
          {sprints.map((sprint, idx) => (
            <Menu.Item className="justify-between" key={sprint.id}>
              <Box className="grid grid-cols-[24px_auto] items-center">
                <SprintsIcon className="h-[1.1rem] w-auto" />
                <Text>{sprint.name}</Text>
              </Box>
              <Flex align="center" gap={2}>
                {/* <CheckIcon className="h-5 w-auto" strokeWidth={2.1} /> */}
                <Text color="muted">{idx}</Text>
              </Flex>
            </Menu.Item>
          ))}
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};
