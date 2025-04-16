import { Box, Flex, Text, Select } from "ui";
import { SunIcon, MoonIcon, SystemIcon } from "icons";
import { useTheme } from "next-themes";
import { SectionHeader } from "@/modules/settings/components";

export const Theming = () => {
  const { theme, setTheme } = useTheme();
  return (
    <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description="Customize the application theme."
        title="Appearance"
      />

      <Box className="p-6">
        <Flex direction="column" gap={6}>
          <Flex className="items-end gap-3 md:items-center" justify="between">
            <Box>
              <Text className="font-medium">Theme</Text>
              <Text className="line-clamp-1" color="muted">
                Select your preferred theme
              </Text>
            </Box>
            <Select
              defaultValue={theme}
              onValueChange={(value) => {
                setTheme(value);
              }}
              value={theme}
            >
              <Select.Trigger className="h-9 w-max shrink-0 px-2 text-[0.9rem] md:text-base">
                <Select.Input />
              </Select.Trigger>
              <Select.Content align="center">
                <Select.Group>
                  <Select.Option className="text-base" value="light">
                    <Flex align="center" gap={2}>
                      <SunIcon className="hidden h-4 md:inline" />
                      Day Mode
                    </Flex>
                  </Select.Option>
                  <Select.Option className="text-base" value="dark">
                    <Flex align="center" gap={2}>
                      <MoonIcon className="hidden h-4 md:inline" />
                      Night Mode
                    </Flex>
                  </Select.Option>
                  <Select.Option className="text-base" value="system">
                    <Flex align="center" gap={2}>
                      <SystemIcon className="hidden h-4 md:inline" />
                      Sync with system
                    </Flex>
                  </Select.Option>
                </Select.Group>
              </Select.Content>
            </Select>
          </Flex>
        </Flex>
      </Box>
    </Box>
  );
};
