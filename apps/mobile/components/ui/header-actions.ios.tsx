import type { HeaderActionsProps } from "./header-actions.types";
import type { SFSymbol } from "expo-symbols";
import { NativeToolbarIcon } from "./native-toolbar-icon";
import {
  Button,
  Host,
  HStack,
  Image,
  Menu,
  RNHostView,
} from "@expo/ui/swift-ui";
import {
  accessibilityLabel,
  accessibilityAddTraits,
  buttonStyle,
  contentShape,
  frame,
  glassEffect,
  shapes,
  tint,
} from "@expo/ui/swift-ui/modifiers";
import { View } from "react-native";
import { NewStoryIcon } from "@/components/icons/new-story";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

export type { HeaderActionsProps } from "./header-actions.types";

function ActionIcon({
  systemName,
  color,
}: {
  systemName: SFSymbol;
  color: string;
}) {
  return (
    <Image
      systemName={systemName}
      size={22}
      color={color}
      modifiers={[
        frame({ width: 44, height: 44 }),
        contentShape(shapes.rectangle()),
      ]}
    />
  );
}

export function HeaderActions({
  onCreate,
  createLabel,
  actions,
  onOptions,
  optionsLabel = "View options",
  menuLabel = "More options",
  menuContent,
  menuSystemImage = "ellipsis",
}: HeaderActionsProps) {
  const { resolvedTheme } = useTheme();
  const foreground = themeColors[resolvedTheme].foreground;
  const hasMenu = Boolean(actions?.length);
  const width = 44 * (1 + Number(Boolean(onOptions)) + Number(hasMenu));

  return (
    <Host
      matchContents
      colorScheme={resolvedTheme}
      style={{ width, height: 44 }}
    >
      <HStack
        spacing={0}
        modifiers={[
          glassEffect({
            glass: { variant: "regular", interactive: true },
            shape: "capsule",
          }),
        ]}
      >
        <Button
          onPress={onCreate}
          modifiers={[buttonStyle("plain"), accessibilityLabel(createLabel)]}
        >
          <HStack
            modifiers={[
              frame({ width: 44, height: 44 }),
              contentShape(shapes.rectangle()),
            ]}
          >
            <RNHostView matchContents>
              <View
                pointerEvents="none"
                style={{
                  width: 44,
                  height: 44,
                  alignItems: "center",
                  justifyContent: "center",
                }}
              >
                <NewStoryIcon size={22} color={foreground} />
              </View>
            </RNHostView>
          </HStack>
        </Button>
        {onOptions ? (
          <Button
            onPress={onOptions}
            modifiers={[buttonStyle("plain"), accessibilityLabel(optionsLabel)]}
          >
            <ActionIcon systemName="ellipsis" color={foreground} />
          </Button>
        ) : null}
        {hasMenu ? (
          <Menu
            modifiers={[buttonStyle("plain"), accessibilityLabel(menuLabel)]}
            label={
              menuContent ? (
                <NativeToolbarIcon>{menuContent}</NativeToolbarIcon>
              ) : (
                <ActionIcon systemName={menuSystemImage} color={foreground} />
              )
            }
          >
            {actions?.map((action) => (
              <Button
                key={action.label}
                label={action.label}
                systemImage={action.selected ? "checkmark" : action.systemImage}
                onPress={action.onPress}
                modifiers={[
                  ...(action.color ? [tint(action.color)] : []),
                  ...(action.selected
                    ? [accessibilityAddTraits(["isSelected"])]
                    : []),
                ]}
              />
            ))}
          </Menu>
        ) : null}
      </HStack>
    </Host>
  );
}
