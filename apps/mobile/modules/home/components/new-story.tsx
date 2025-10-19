import React from "react";
import { View, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
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
      <Pressable
        onPress={handleNewStory}
        style={{
          width: 60,
          height: 60,
          borderRadius: 30,
          backgroundColor: "rgba(255, 255, 255, 0.8)",
          justifyContent: "center",
          alignItems: "center",
          shadowColor: "#000",
          shadowOffset: {
            width: 0,
            height: 4,
          },
          shadowOpacity: 0.2,
          shadowRadius: 8,
          elevation: 6,
        }}
      >
        <Ionicons name="add" size={23} color="black" />
      </Pressable>
    </View>
  );
};
