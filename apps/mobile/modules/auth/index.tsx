import { Pressable, View } from "react-native";
import { Image } from "expo-image";
import { Button, Col, SafeContainer, Text } from "@/components/ui";
import { useAuthStore } from "@/store";
import { Logo } from "@/components/icons";
import { themeColors } from "@/constants/colors";
import * as WebBrowser from "expo-web-browser";
import {
  authenticateWithCode,
  beginMobileSignIn,
  MOBILE_REDIRECT_URI,
} from "@/lib/actions/auth";
import { clearSignInTransaction } from "@/lib/auth";
import { useEffect, useRef, useState } from "react";
import { useLocalSearchParams } from "expo-router";
import { useTheme } from "@/hooks";
import { openPublicPage, publicPages } from "@/lib/public-pages";
const lightMesh = require("@/assets/images/mesh.webp");
const darkMesh = require("@/assets/images/mesh-dark.webp");

export const Auth = () => {
  const { resolvedTheme } = useTheme();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const sessionError = useAuthStore((state) => state.sessionError);
  const loadAuthData = useAuthStore((state) => state.loadAuthData);
  const setAuthData = useAuthStore((state) => state.setAuthData);
  const callback = useLocalSearchParams<{ code?: string; state?: string }>();
  const handledCallback = useRef<string | null>(null);

  useEffect(() => {
    if (typeof callback.code !== "string" || typeof callback.state !== "string")
      return;
    const url = new URL(MOBILE_REDIRECT_URI);
    url.searchParams.set("code", callback.code);
    url.searchParams.set("state", callback.state);
    if (handledCallback.current === url.toString()) return;
    handledCallback.current = url.toString();
    setLoading(true);
    setError(null);
    void authenticateWithCode(url.toString())
      .then(setAuthData)
      .catch((cause: unknown) => {
        setError(
          cause instanceof Error
            ? cause.message
            : "Unable to complete sign-in. Please try again.",
        );
      })
      .finally(() => {
        setLoading(false);
      });
  }, [callback.code, callback.state, setAuthData]);

  const handleGetStarted = async () => {
    if (loading) return;
    try {
      setLoading(true);
      setError(null);
      const result = await WebBrowser.openAuthSessionAsync(
        await beginMobileSignIn(),
        MOBILE_REDIRECT_URI,
      );

      if (result.type === "success" && result.url) {
        const res = await authenticateWithCode(result.url);
        await setAuthData(res);
      }
    } catch (error) {
      setError(
        error instanceof Error
          ? error.message
          : "Unable to sign in. Please try again.",
      );
    }
    await clearSignInTransaction().catch(() => {
      setError("Unable to clear the sign-in request. Please try again.");
    });
    setLoading(false);
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
        <Logo height={30} color={themeColors[resolvedTheme].foreground} />
        <Col>
          {error || sessionError ? (
            <Text accessibilityRole="alert" align="center" className="mb-4">
              {error ?? sessionError}
            </Text>
          ) : null}
          <Text
            className="mb-6 uppercase text-[14px] tracking-wider"
            fontSize="sm"
          >
            [Built for builders]
          </Text>
          <Text fontSize="4xl" fontWeight="bold">
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
            Continue with email
          </Button>
          {sessionError ? (
            <Button
              onPress={() => {
                void loadAuthData();
              }}
              disabled={loading}
              className="mt-3"
            >
              Retry connection
            </Button>
          ) : null}
          <View
            style={{
              flexDirection: "row",
              justifyContent: "center",
              flexWrap: "wrap",
              marginTop: 8,
            }}
          >
            {(["privacy", "terms"] as const).map((page) => (
              <Pressable
                key={page}
                accessibilityRole="link"
                onPress={() => openPublicPage(page)}
                style={({ pressed }) => ({
                  minHeight: 44,
                  paddingHorizontal: 12,
                  justifyContent: "center",
                  opacity: pressed ? 0.6 : 1,
                })}
              >
                <Text fontSize="sm" decoration="underline">
                  {publicPages[page].label}
                </Text>
              </Pressable>
            ))}
          </View>
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
