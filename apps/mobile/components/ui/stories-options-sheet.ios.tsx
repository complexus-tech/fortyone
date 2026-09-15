import type { StoriesOptionsSheetProps } from "./stories-options-sheet.shared";
import { useState } from "react";
import { useWindowDimensions, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import {
  Button,
  Divider,
  HStack,
  Image,
  Menu,
  RNHostView,
  ScrollView,
  Spacer,
  Text,
  VStack,
} from "@expo/ui/swift-ui";
import {
  accessibilityLabel,
  buttonStyle,
  contentShape,
  font,
  foregroundStyle,
  frame,
  onGeometryChange,
  padding,
  shapes,
} from "@expo/ui/swift-ui/modifiers";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import { BottomSheetModal } from "./bottom-sheet-modal";
import {
  getStoriesOptionRows,
  StoriesDisplayOptions,
} from "./stories-options-sheet.shared";

const SHEET_GUTTER = 20;
const SHEET_TOP_PADDING = 24;
const CONTENT_GAP = 24;
const CONTENT_BOTTOM_PADDING = 20;

export const StoriesOptionsSheet = (props: StoriesOptionsSheetProps) => {
  const { isOpened, setIsOpened } = props;
  const { resolvedTheme } = useTheme();
  const { width: windowWidth, height: windowHeight } = useWindowDimensions();
  const insets = useSafeAreaInsets();
  const availableWidth = Math.max(1, windowWidth - SHEET_GUTTER * 2);
  const [contentSize, setContentSize] = useState({
    width: availableWidth,
    menuHeight: 0,
    propertiesHeight: 0,
  });
  const contentHeight =
    contentSize.menuHeight > 0 && contentSize.propertiesHeight > 0
      ? Math.ceil(
          contentSize.menuHeight +
            CONTENT_GAP +
            contentSize.propertiesHeight +
            CONTENT_BOTTOM_PADDING,
        )
      : 360;
  const maximumHeight = Math.max(
    1,
    (windowHeight - insets.top - insets.bottom) * 0.85 - SHEET_TOP_PADDING,
  );
  const rows = getStoriesOptionRows(props);
  const isDark = resolvedTheme === "dark";
  const foreground = themeColors[isDark ? "dark" : "light"].foreground;
  const muted = themeColors[isDark ? "dark" : "light"].textMuted;

  return (
    <BottomSheetModal
      nativeContent
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
      spacing={0}
      padding={{
        leading: SHEET_GUTTER,
        trailing: SHEET_GUTTER,
        top: SHEET_TOP_PADDING,
        bottom: 0,
      }}
    >
      <ScrollView
        showsIndicators={contentHeight > maximumHeight}
        modifiers={[
          frame({ minWidth: 0, maxWidth: availableWidth }),
          frame({ height: Math.min(contentHeight, maximumHeight) }),
          onGeometryChange(({ width }) => {
            if (width <= 0) return;
            setContentSize((current) =>
              current.width === width ? current : { ...current, width },
            );
          }),
        ]}
      >
        <VStack
          spacing={CONTENT_GAP}
          alignment="leading"
          modifiers={[padding({ bottom: CONTENT_BOTTOM_PADDING })]}
        >
          <VStack
            spacing={0}
            modifiers={[
              onGeometryChange(({ height }) => {
                if (height <= 0) return;
                setContentSize((current) =>
                  current.menuHeight === height
                    ? current
                    : { ...current, menuHeight: height },
                );
              }),
            ]}
          >
            {rows.map((row, index) => (
              <VStack key={row.id} spacing={0}>
                {index > 0 ? <Divider /> : null}
                <Menu
                  modifiers={[
                    buttonStyle("plain"),
                    accessibilityLabel(`${row.label}, ${row.value}`),
                  ]}
                  label={
                    <HStack
                      spacing={16}
                      modifiers={[
                        padding({ vertical: 14 }),
                        frame({ minHeight: 52 }),
                        contentShape(shapes.rectangle()),
                      ]}
                    >
                      <Text
                        modifiers={[
                          font({ textStyle: "callout", weight: "regular" }),
                          foregroundStyle(foreground),
                        ]}
                      >
                        {row.label}
                      </Text>
                      <Spacer />
                      <HStack spacing={8}>
                        <Text
                          modifiers={[
                            font({ textStyle: "callout", weight: "regular" }),
                            foregroundStyle(muted),
                          ]}
                        >
                          {row.value}
                        </Text>
                        <Image
                          systemName="chevron.up.chevron.down"
                          size={12}
                          color={muted}
                        />
                      </HStack>
                    </HStack>
                  }
                >
                  {row.actions.map((action) => (
                    <Button
                      key={action.label}
                      label={action.label}
                      systemImage={action.selected ? "checkmark" : undefined}
                      onPress={action.onPress}
                    />
                  ))}
                </Menu>
              </VStack>
            ))}
          </VStack>
          <RNHostView matchContents>
            <View
              style={{ width: Math.min(contentSize.width, availableWidth) }}
              onLayout={({ nativeEvent: { layout } }) => {
                if (layout.height <= 0) return;
                setContentSize((current) =>
                  current.propertiesHeight === layout.height
                    ? current
                    : { ...current, propertiesHeight: layout.height },
                );
              }}
            >
              <StoriesDisplayOptions {...props} />
            </View>
          </RNHostView>
        </VStack>
      </ScrollView>
    </BottomSheetModal>
  );
};
