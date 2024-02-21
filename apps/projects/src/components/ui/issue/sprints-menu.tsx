import { Box, Button, Flex, Menu, Text } from "ui";
import { CheckIcon, PlusIcon, SprintsIcon } from "@/components/icons";

export const SprintsMenu = () => {
  const sprints = ["Complains", "Clients", "Audit"];
  return (
    <Menu>
      <Menu.Button>
        <Button
          color="tertiary"
          leftIcon={<PlusIcon className="h-5 w-auto" />}
          size="md"
          variant="naked"
        >
          Add sprint
        </Button>
      </Menu.Button>
      <Menu.Items align="center" className="w-64">
        <Menu.Group className="px-4">
          <Menu.Input autoFocus placeholder="Select sprint..." />
        </Menu.Group>
        <Menu.Separator className="my-2" />
        <Menu.Group>
          {sprints.map((sprint, idx) => (
            <Menu.Item className="justify-between" key={sprint}>
              <Box className="grid grid-cols-[24px_auto] items-center">
                <SprintsIcon className="h-[1.1rem] w-auto" />
                <Text>{sprint}</Text>
              </Box>
              <Flex align="center" gap={2}>
                {sprint === "io" && (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                )}
                <Text color="muted">{idx}</Text>
              </Flex>
            </Menu.Item>
          ))}
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};
