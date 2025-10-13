import { View } from "react-native";
import { Image } from "expo-image";
import { Button, Col, SafeContainer, Text } from "@/components/ui";
import { useAuthStore } from "@/store";
import { Logo } from "@/components/icons";
import { colors } from "@/constants";
import * as WebBrowser from "expo-web-browser";
import { authenticateWithToken } from "@/lib/actions/auth";
import { useState } from "react";

export const Auth = () => {
  const [loading, setLoading] = useState(false);
  const setAuthData = useAuthStore((state) => state.setAuthData);

  const handleGetStarted = async () => {
    try {
      // Open the landing app login page in an in-app browser
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

        console.log(res);
        setAuthData(res.token, res.workspace);

        // setAuthData(
        //   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI4YTc5ODExMi05MGZlLTQ5NWUtOWYxYy1mMzY2NTVlM2Q4YWIiLCJleHAiOjE3NjI4ODMyMDMsIm5iZiI6MTc1OTQyNzIwMywiaWF0IjoxNzU5NDI3MjAzfQ.K05W85tEEWQ5dFqu7bgXjjowkk_zYowwKSJ_VMXR7_o",
        //   "complexus"
        // );
      }
    } catch (error) {
      console.error("Authentication error:", error);
      // TODO: Show error message to user
    }
  };

  return (
    <View
      style={{
        flex: 1,
      }}
    >
      <Image
        source={require("@/assets/images/mesh.webp")}
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
        <Logo height={30} color={colors.dark.DEFAULT} />
        <Col>
          <Text
            className="mb-6 uppercase text-[14px] tracking-wider"
            fontSize="sm"
            color="black"
          >
            [Built for builders]
          </Text>
          <Text fontSize="4xl" fontWeight="semibold" color="black">
            Plan, track, deliver with the project management tool your team will
            love.
          </Text>
        </Col>
        <Col>
          <Button
            size="lg"
            rounded="lg"
            className="w-full bg-dark border-dark"
            onPress={handleGetStarted}
            loading={loading}
          >
            Get Started
          </Button>
          <Text
            align="center"
            className="mt-4 opacity-80 text-[15px]"
            color="black"
          >
            © {new Date().getFullYear()} • Product of Complexus LLC • All
            Rights Reserved.
          </Text>
        </Col>
      </SafeContainer>
    </View>
  );
};
