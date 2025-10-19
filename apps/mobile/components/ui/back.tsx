import React from "react";
import { useRouter } from "expo-router";
import { Pressable } from "react-native";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";
import { Ionicons } from "@expo/vector-icons";

export const Back = () => {
  const { resolvedTheme } = useTheme();
  const router = useRouter();
  const canGoBack = router.canGoBack();

  const handleBack = () => {
    if (canGoBack) {
      router.back();
    } else {
      router.replace("/");
    }
  };

  return (
    <Pressable
      onPress={handleBack}
      style={{
        width: 40,
        height: 40,
        borderRadius: 20,
        backgroundColor:
          resolvedTheme === "light"
            ? "rgba(255, 255, 255, 0.8)"
            : "rgba(0, 0, 0, 0.8)",
        justifyContent: "center",
        alignItems: "center",
        shadowColor: "#000",
        shadowOffset: {
          width: 0,
          height: 2,
        },
        shadowOpacity: 0.1,
        shadowRadius: 4,
        elevation: 3,
      }}
    >
      <Ionicons
        name="chevron-back"
        size={20}
        color={resolvedTheme === "light" ? colors.dark[50] : colors.gray[200]}
      />
    </Pressable>
  );
};
