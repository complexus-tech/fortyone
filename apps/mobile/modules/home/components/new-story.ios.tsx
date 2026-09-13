import React from "react";
import { View } from "react-native";
import { Host, Button, Image } from "@expo/ui/swift-ui";
import {
  frame,
  glassEffect,
  accessibilityLabel,
  buttonStyle,
} from "@expo/ui/swift-ui/modifiers";
import { useRouter } from "expo-router";

export const NewStoryButton = () => {
  const router = useRouter();
  const handleNewStory = () => {
    router.push("/new");
  };
  return (
    <View
      style={{
        position: "absolute",
        bottom: 100,
        right: 24,
        width: 60,
        height: 60,
      }}
    >
      <Host
        matchContents
        style={{
          width: 60,
          height: 60,
        }}
      >
        <Button
          modifiers={[
            accessibilityLabel("Create task"),
            buttonStyle("plain"),
            frame({ width: 60, height: 60 }),
            glassEffect({
              glass: {
                interactive: true,
                variant: "regular",
              },
            }),
          ]}
          onPress={handleNewStory}
        >
          <Image systemName="plus" size={23} />
        </Button>
      </Host>
    </View>
  );
};
