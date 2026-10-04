import { useId, useSyncExternalStore } from "react";
import { Box, Flex, Text, Select, Switch } from "ui";
import { SunIcon, MoonIcon, SystemIcon } from "icons";
import { useTheme } from "next-themes";
import { SectionHeader } from "@/modules/settings/components";
import { useTerminology } from "@/hooks";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { useUpdateAutomationPreferencesMutation } from "@/lib/hooks/users/update-auto-preferences";

const subscribeToHydration = () => () => {};
const getClientSnapshot = () => true;
const getServerSnapshot = () => false;

export const Theming = () => {
  const id = useId();
  const { theme, setTheme } = useTheme();
  const isHydrated = useSyncExternalStore(
    subscribeToHydration,
    getClientSnapshot,
    getServerSnapshot,
  );
  // next-themes reads browser storage before hydration; keep Radix's initial
  // value consistent with the server, then display the saved preference.
  const selectedTheme = isHydrated ? theme ?? "system" : "system";
  const { getTermDisplay } = useTerminology();
  const { data: preferences } = useAutomationPreferences();
  const { mutate: updatePreferences } =
    useUpdateAutomationPreferencesMutation();
  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        description="Customize how the application looks and behaves."
        title="Appearance & Behavior"
      />

      <Box className="p-6">
        <Flex direction="column" gap={6}>
          <Flex
            className="flex-col items-start gap-3 md:flex-row md:items-center"
            justify="between"
          >
            <Box className="min-w-0">
              <Text className="font-medium" id={`${id}-appearance-label`}>
                Appearance
              </Text>
              <Text color="muted" id={`${id}-appearance-description`}>
                Select your preferred theme
              </Text>
            </Box>
            <Select
              onValueChange={(value) => {
                setTheme(value);
              }}
              value={selectedTheme}
            >
              <Select.Trigger
                aria-describedby={`${id}-appearance-description`}
                aria-labelledby={`${id}-appearance-label`}
                className="w-max shrink-0 text-base"
              >
                <Select.Input />
              </Select.Trigger>
              <Select.Content
                align="center"
                className="w-max max-w-[calc(100vw-2rem)]"
              >
                <Select.Group>
                  <Select.Option className="text-base" value="light">
                    <Flex
                      align="center"
                      as="span"
                      className="inline-flex"
                      gap={2}
                    >
                      <SunIcon
                        className="h-5 w-auto shrink-0"
                        strokeWidth={1.5}
                      />
                      Day Mode
                    </Flex>
                  </Select.Option>
                  <Select.Option className="text-base" value="dark">
                    <Flex
                      align="center"
                      as="span"
                      className="inline-flex"
                      gap={2}
                    >
                      <MoonIcon
                        className="h-5 w-auto shrink-0"
                        strokeWidth={1.5}
                      />
                      Night Mode
                    </Flex>
                  </Select.Option>
                  <Select.Option className="text-base" value="system">
                    <Flex
                      align="center"
                      as="span"
                      className="inline-flex"
                      gap={2}
                    >
                      <SystemIcon
                        className="h-5 w-auto shrink-0"
                        strokeWidth={1.5}
                      />
                      Sync with system
                    </Flex>
                  </Select.Option>
                </Select.Group>
              </Select.Content>
            </Select>
          </Flex>
          <Flex align="center" gap={2} justify="between">
            <Box>
              <Text className="font-medium" id={`${id}-dialog-label`}>
                On {getTermDisplay("storyTerm")} click, open in dialog
              </Text>
              <Text
                className="line-clamp-2"
                color="muted"
                id={`${id}-dialog-description`}
              >
                After clicking a {getTermDisplay("storyTerm")}, it opens in a
                dialog
              </Text>
            </Box>
            <Switch
              aria-describedby={`${id}-dialog-description`}
              aria-labelledby={`${id}-dialog-label`}
              checked={preferences?.openStoryInDialog}
              className="shrink-0"
              name="openStoryInDialog"
              onCheckedChange={(checked) => {
                updatePreferences({ openStoryInDialog: checked });
              }}
            />
          </Flex>
        </Flex>
      </Box>
    </Box>
  );
};
