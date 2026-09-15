import type { ComponentProps } from "react";
import type { PressableProps } from "react-native";
import type { SwipeableMethods } from "react-native-gesture-handler/ReanimatedSwipeable";
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  ActionSheetIOS,
  Alert,
  Platform,
  Pressable,
  View,
  useColorScheme,
  useWindowDimensions,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import ReanimatedSwipeable from "react-native-gesture-handler/ReanimatedSwipeable";
import { ReduceMotion } from "react-native-reanimated";
import { themeColors } from "@/constants/colors";
import { Text } from "./Text";
import { createSwipeRowInteraction } from "./swipeable-row-interaction";
import {
  captureSwipeAction,
  resolveSwipeAction,
  type SwipeActionToken,
} from "./swipeable-row-actions";

export type RowAction = {
  id?: string;
  label: string;
  accessibilityLabel: string;
  icon: ComponentProps<typeof Ionicons>["name"];
  backgroundColor: string;
  foregroundColor: string;
  onPress: () => void;
  disabled?: boolean;
  destructive?: boolean;
};

type SwipeableRowProps = PressableProps & {
  children: React.ReactNode;
  className?: string;
  actionAppearance?: "inset" | "flush";
  /** Identifies the row and session to invalidate choices in an open native menu. */
  actionScope?: string;
} & (
    | { action?: RowAction; actions?: never }
    // Two actions plus Cancel fit both native platforms' action menus.
    | { action?: never; actions: readonly [RowAction, RowAction] }
  );

// Only one action rail remains open, including when moving between lists.
// This holds an ephemeral close callback, never row or account data.
let closeOpenRow: (() => void) | undefined;

export function SwipeableRow(props: SwipeableRowProps) {
  return <SwipeableRowContent key={props.actionScope ?? "row"} {...props} />;
}

function SwipeableRowContent({
  action,
  actions,
  actionAppearance = "inset",
  actionScope,
  children,
  accessibilityHint,
  accessibilityActions,
  onAccessibilityAction,
  onLongPress,
  disabled,
  onPress,
  onTouchStart,
  onTouchMove,
  ...props
}: SwipeableRowProps) {
  const swipeable = useRef<SwipeableMethods>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [interaction] = useState(createSwipeRowInteraction);
  const colorScheme = useColorScheme();
  const backgroundColor =
    themeColors[colorScheme === "dark" ? "dark" : "light"].background;
  const { fontScale } = useWindowDimensions();
  const rowActions = useMemo(
    () => actions ?? (action ? [action] : []),
    [actions, action],
  );
  const currentActions = useRef({
    actions: rowActions,
    disabled,
    active: true,
    scope: actionScope,
  });
  useLayoutEffect(() => {
    currentActions.current = {
      actions: rowActions,
      disabled,
      active: true,
      scope: actionScope,
    };
    return () => {
      currentActions.current.active = false;
    };
  }, [rowActions, disabled, actionScope]);
  const close = useCallback(() => {
    if (!swipeable.current) {
      interaction.didClose();
      return;
    }
    if (!interaction.requestClose()) return;
    swipeable.current.close();
  }, [interaction]);
  useEffect(
    () => () => {
      if (closeOpenRow === close) closeOpenRow = undefined;
    },
    [close],
  );

  const performAction = (token: SwipeActionToken) => {
    interaction.cancelPress();
    close();
    resolveSwipeAction(currentActions.current, token)?.onPress();
  };
  const enabledActions = rowActions.filter((item) => !item.disabled);
  const entries = enabledActions.map((item) => ({
    action: item,
    token: captureSwipeAction(item, actionScope),
    name: action
      ? "quickAction"
      : `quickAction:${encodeURIComponent(item.id ?? item.accessibilityLabel)}`,
  }));
  const showActions = () => {
    if (entries.length === 0 || disabled) return;
    interaction.cancelPress();
    close();
    if (Platform.OS === "ios") {
      ActionSheetIOS.showActionSheetWithOptions(
        {
          title: props.accessibilityLabel,
          options: [
            ...entries.map((entry) => entry.action.accessibilityLabel),
            "Cancel",
          ],
          cancelButtonIndex: entries.length,
          destructiveButtonIndex: entries.flatMap((entry, index) =>
            entry.action.destructive ? [index] : [],
          ),
        },
        (index) => {
          const selected = entries[index];
          if (selected) performAction(selected.token);
        },
      );
    } else {
      Alert.alert(props.accessibilityLabel ?? "Actions", undefined, [
        { text: "Cancel", style: "cancel" },
        ...entries.map((entry) => ({
          text: entry.action.accessibilityLabel,
          style: entry.action.destructive
            ? ("destructive" as const)
            : ("default" as const),
          onPress: () => performAction(entry.token),
        })),
      ]);
    }
  };
  const hasActions = rowActions.length > 0;
  const content = (
    <Pressable
      {...props}
      disabled={disabled}
      onTouchStart={(event) => {
        // onPressIn can repeat when the same finger re-enters the press rect.
        // Only an actual new touch is allowed to reset a cancelled swipe.
        interaction.touchStarted({
          x: event.nativeEvent.pageX,
          y: event.nativeEvent.pageY,
        });
        if (event.nativeEvent.touches.length > 1) interaction.cancelPress();
        onTouchStart?.(event);
      }}
      onTouchMove={(event) => {
        interaction.pressMoved({
          x: event.nativeEvent.pageX,
          y: event.nativeEvent.pageY,
        });
        onTouchMove?.(event);
      }}
      onPress={(event) => {
        // Pressability also forwards non-touch accessibility/keyboard clicks
        // directly to onPress, without a touch-start event or touch payload.
        const source = "touches" in event.nativeEvent ? "touch" : "activation";
        if (!interaction.canOpenDetails(source)) {
          close();
          return;
        }
        onPress?.(event);
      }}
      onLongPress={entries.length > 0 ? showActions : onLongPress}
      accessibilityHint={
        entries.length > 0
          ? `${accessibilityHint ?? "Open details"}. Swipe left or long press for actions.`
          : accessibilityHint
      }
      accessibilityActions={[
        ...(accessibilityActions ?? []),
        ...(!disabled
          ? entries.map((entry) => ({
              name: entry.name,
              label: entry.action.accessibilityLabel,
            }))
          : []),
      ]}
      onAccessibilityAction={(event) => {
        const selected = entries.find(
          (entry) => entry.name === event.nativeEvent.actionName,
        );
        if (selected) performAction(selected.token);
        else onAccessibilityAction?.(event);
      }}
    >
      {children}
    </Pressable>
  );

  if (!hasActions) return content;
  const flush = actionAppearance === "flush";
  const actionWidth = flush ? Math.max(84, Math.min(120, 84 * fontScale)) : 88;
  return (
    <ReanimatedSwipeable
      ref={swipeable}
      enabled={!disabled && entries.length > 0}
      friction={1}
      rightThreshold={44}
      dragOffsetFromRightEdge={20}
      overshootRight={false}
      overshootLeft={false}
      enableTrackpadTwoFingerGesture
      animationOptions={{ reduceMotion: ReduceMotion.System }}
      childrenContainerStyle={{ backgroundColor }}
      onSwipeableOpenStartDrag={() => {
        interaction.swipeStarted();
        if (closeOpenRow !== close) closeOpenRow?.();
        closeOpenRow = close;
      }}
      onSwipeableCloseStartDrag={() => interaction.swipeStarted()}
      onSwipeableWillOpen={() => {
        interaction.willOpen();
        setIsOpen(true);
        if (closeOpenRow !== close) closeOpenRow?.();
        closeOpenRow = close;
      }}
      onSwipeableWillClose={() => interaction.willClose()}
      onSwipeableClose={() => {
        interaction.didClose();
        setIsOpen(false);
        if (closeOpenRow === close) closeOpenRow = undefined;
      }}
      renderRightActions={() => (
        <View
          accessibilityElementsHidden={!isOpen}
          importantForAccessibility={isOpen ? "auto" : "no-hide-descendants"}
          style={{
            flexDirection: "row",
            width: actionWidth * rowActions.length + (flush ? 0 : 8),
            paddingVertical: flush ? 0 : 4,
            paddingRight: flush ? 0 : 8,
          }}
        >
          {rowActions.map((item) => (
            <Pressable
              key={item.id ?? item.accessibilityLabel}
              accessibilityRole="button"
              accessibilityLabel={item.accessibilityLabel}
              accessibilityState={{
                disabled: Boolean(disabled || item.disabled),
              }}
              disabled={disabled || item.disabled}
              onPress={() =>
                performAction(captureSwipeAction(item, actionScope))
              }
              style={({ pressed }) => ({
                width: actionWidth,
                minHeight: 44,
                borderRadius: flush ? 0 : 14,
                gap: 3,
                alignItems: "center",
                justifyContent: "center",
                backgroundColor: item.backgroundColor,
                opacity: pressed || disabled || item.disabled ? 0.6 : 1,
              })}
            >
              <Ionicons
                accessible={false}
                name={item.icon}
                size={22}
                color={item.foregroundColor}
              />
              <Text
                numberOfLines={1}
                style={{
                  fontSize: flush ? 13 : 12,
                  lineHeight: flush ? 18 : 16,
                  fontWeight: "500",
                  color: item.foregroundColor,
                }}
              >
                {item.label}
              </Text>
            </Pressable>
          ))}
        </View>
      )}
    >
      {content}
    </ReanimatedSwipeable>
  );
}
