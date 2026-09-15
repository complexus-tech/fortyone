import { useState } from "react";
import { View, useWindowDimensions } from "react-native";
import { useNavigation, useRouter } from "expo-router";
import {
  BottomSheet,
  Group,
  Host,
  RNHostView,
  ScrollView,
} from "@expo/ui/swift-ui";
import {
  frame,
  onGeometryChange,
  presentationDragIndicator,
} from "@expo/ui/swift-ui/modifiers";
import { useTheme } from "@/hooks/theme";
import { useAuthStore } from "@/store/auth";
import { Form } from "./components/form";
import { Header } from "./components/header";

export const Settings = () => {
  const router = useRouter();
  const navigation = useNavigation();
  const { resolvedTheme } = useTheme();
  const { width: windowWidth, height: windowHeight } = useWindowDimensions();
  const [isPresented, setIsPresented] = useState(true);
  const sessionEpoch = useAuthStore((state) => state.sessionEpoch);
  const [openingSession] = useState(sessionEpoch);
  const [contentSize, setContentSize] = useState({
    width: windowWidth,
    height: 560,
  });
  const maximumHeight = windowHeight * 0.82;

  const finishDismissal = () => {
    const session = useAuthStore.getState();
    // Authentication and workspace changes can remove this route while the
    // native sheet is dismissing. Never pop the replacement session's screen.
    if (
      session.isLoading ||
      !session.isAuthenticated ||
      session.sessionEpoch !== openingSession ||
      !navigation.isFocused()
    ) {
      return;
    }
    if (router.canGoBack()) router.back();
    else router.replace("/");
  };

  return (
    <View style={{ flex: 1, backgroundColor: "transparent" }}>
      <Host
        matchContents
        colorScheme={resolvedTheme}
        style={{ position: "absolute" }}
      >
        <BottomSheet
          isPresented={isPresented}
          onIsPresentedChange={setIsPresented}
          onDismiss={finishDismissal}
          fitToContents
        >
          {/* Leave the SwiftUI sheet's system material unobscured, just like
              the native My Work options sheet. */}
          <Group modifiers={[presentationDragIndicator("visible")]}>
            <ScrollView
              showsIndicators={contentSize.height > maximumHeight}
              modifiers={[
                // Give the viewport its own flexible width before measuring it.
                // A self-sizing RN child cannot bootstrap an empty ScrollView.
                frame({ minWidth: 0, maxWidth: windowWidth }),
                frame({
                  height: Math.min(contentSize.height, maximumHeight),
                }),
                onGeometryChange(({ width }) => {
                  if (width <= 0) return;
                  setContentSize((current) =>
                    current.width === width ? current : { ...current, width },
                  );
                }),
              ]}
            >
              <RNHostView matchContents>
                <View
                  style={{
                    width: Math.min(contentSize.width, windowWidth),
                    paddingTop: 20,
                    paddingBottom: 24,
                  }}
                  onLayout={({ nativeEvent: { layout } }) => {
                    if (layout.height <= 0) return;
                    setContentSize((current) =>
                      current.height === layout.height
                        ? current
                        : { ...current, height: layout.height },
                    );
                  }}
                >
                  <Header onClose={() => setIsPresented(false)} />
                  <Form />
                </View>
              </RNHostView>
            </ScrollView>
          </Group>
        </BottomSheet>
      </Host>
    </View>
  );
};
