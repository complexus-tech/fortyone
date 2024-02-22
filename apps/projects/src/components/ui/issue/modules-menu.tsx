import { Box, Button, Flex, Menu, Text } from "ui";
import { CheckIcon, ModulesIcon, PlusIcon } from "icons";

export const ModulesMenu = () => {
  const modules = ["Complains", "Clients", "Audit"];
  return (
    <Menu>
      <Menu.Button>
        <Button
          color="tertiary"
          leftIcon={<PlusIcon className="h-5 w-auto" />}
          size="md"
          variant="naked"
        >
          Add module
        </Button>
      </Menu.Button>
      <Menu.Items align="center" className="w-64">
        <Menu.Group className="px-4">
          <Menu.Input autoFocus placeholder="Select module..." />
        </Menu.Group>
        <Menu.Separator className="my-2" />
        <Menu.Group>
          {modules.map((mod, idx) => (
            <Menu.Item className="justify-between" key={mod}>
              <Box className="grid grid-cols-[24px_auto] items-center">
                <ModulesIcon className="h-[1.1rem] w-auto" />
                <Text>{mod}</Text>
              </Box>
              <Flex align="center" gap={2}>
                {mod === "io" && (
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
