import { View } from "react-native";
import { Image } from "expo-image";
import { Button, Col, SafeContainer, Text } from "@/components/ui";
import { useAuthStore } from "@/store";
import { Logo } from "@/components/icons";
import { colors } from "@/constants";
import * as WebBrowser from "expo-web-browser";
import { authenticateWithToken } from "@/lib/actions/auth";
import { useState } from "react";
import { useTheme } from "@/hooks";
const lightMesh = require("@/assets/images/mesh.webp");
const darkMesh = require("@/assets/images/mesh-dark.webp");

export const Auth = () => {
  const { resolvedTheme } = useTheme();
  const [loading, setLoading] = useState(false);
  const setAuthData = useAuthStore((state) => state.setAuthData);

  const handleGetStarted = async () => {
    try {
      const result = await WebBrowser.openAuthSessionAsync(
        "https://www.fortyone.app/login?mobile=true",
        "fortyone://login"
      );

      if (result.type === "success" && result.url) {
        const url = new URL(result.url);
        const code = url.searchParams.get("code") ?? "";
        const email = url.searchParams.get("email") ?? "";
        setLoading(true);
        const res = await authenticateWithToken(email, code);
        setLoading(false);
        setAuthData(res.token, res.workspace);
      }
    } catch (error) {
      console.error("Authentication error:", error);
    }
  };

  return (
    <View
      style={{
        flex: 1,
      }}
    >
      <Image
        source={resolvedTheme === "dark" ? darkMesh : lightMesh}
        style={{
          position: "absolute",
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          width: "100%",
          height: "100%",
        }}
        contentFit="cover"
        contentPosition="center"
      />
      <SafeContainer
        style={{
          backgroundColor: "transparent",
          justifyContent: "space-between",
          flex: 1,
          paddingBottom: 50,
          paddingTop: 5,
        }}
      >
        <Logo
          height={30}
          color={resolvedTheme === "dark" ? colors.white : colors.black}
        />
        <Col>
          <Text
            className="mb-6 uppercase text-[14px] tracking-wider"
            fontSize="sm"
          >
            [Built for builders]
          </Text>
          <Text fontSize="4xl" fontWeight="semibold">
            Plan, track, deliver with the project management tool your team will
            love.
          </Text>
        </Col>
        <Col>
          <Button
            size="lg"
            rounded="lg"
            className="w-full"
            color="invert"
            onPress={handleGetStarted}
            loading={loading}
          >
            Get Started
          </Button>
          <Text
            align="center"
            className="mt-4 opacity-80 text-[15px] dark:opacity-100"
          >
            © {new Date().getFullYear()} • Product of Complexus LLC • All
            Rights Reserved.
          </Text>
        </Col>
      </SafeContainer>
    </View>
  );
};
