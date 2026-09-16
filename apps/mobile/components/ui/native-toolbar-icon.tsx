import type { ReactNode } from "react";
import { View } from "react-native";
import { HStack, RNHostView } from "@expo/ui/swift-ui";
import { contentShape, frame, shapes } from "@expo/ui/swift-ui/modifiers";

/** Hosts a custom glyph inside a native toolbar's 44-point hit area. */
export function NativeToolbarIcon({ children }: { children: ReactNode }) {
  return (
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
          {children}
        </View>
      </RNHostView>
    </HStack>
  );
}
