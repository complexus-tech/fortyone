"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Input, Text, Button, Flex } from "ui";
import Link from "next/link";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { Logo, GoogleIcon, MicrosoftIcon } from "@/components/ui";
import { OTPInput } from "@/components/ui/otp-input";
import { requestMagicEmail } from "@/lib/actions/request-magic-email";
import { signInWithGoogle, signInWithMicrosoft } from "@/lib/actions/sign-in";
import { getSafeCallbackUrl, withCallbackUrl } from "@/utils/callback-url";
import { isMobileAuthFlow } from "@/lib/mobile-auth";
import { getSignInErrorMessage } from "./errors";

const COPYRIGHT_NOTICE =
  "\u00a9 2026 \u2022 Product of Complexus LLC \u2022 All Rights Reserved.";

export const AuthLayout = ({
  page,
  errorMessage,
  callbackUrl,
  isMobileApp: mobileApp = false,
  accountDeleted = false,
  cleanupPending = false,
}: {
  page: "login" | "signup";
  errorMessage?: string;
  callbackUrl?: string;
  isMobileApp?: boolean;
  accountDeleted?: boolean;
  cleanupPending?: boolean;
}) => {
  const [loading, setLoading] = useState(false);
  const [email, setEmail] = useState("");
  const [isSent, setIsSent] = useState(false);
  const [isTouched, setIsTouched] = useState(false);
  const [otp, setOtp] = useState("");
  const [otpLoading, setOtpLoading] = useState(false);
  const router = useRouter();
  const safeCallbackUrl = getSafeCallbackUrl(callbackUrl);
  const isMobileApp = mobileApp || isMobileAuthFlow(safeCallbackUrl);
  const signInError = getSignInErrorMessage(errorMessage);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    const result = await requestMagicEmail(
      email,
      isMobileApp,
      safeCallbackUrl,
    ).finally(() => {
      setLoading(false);
    });

    if (result.error.message) {
      toast.error(
        isMobileApp
          ? "Failed to send sign-in code"
          : "Failed to send magic link",
        {
          description: result.error.message,
        },
      );
      return;
    }

    setIsSent(true);
  };

  const handleOTPSubmit = async () => {
    const url = withCallbackUrl(
      `/verify/${encodeURIComponent(email)}/${encodeURIComponent(otp)}${isMobileApp ? "?mobileApp=true" : ""}`,
      safeCallbackUrl,
    );

    if (otp.length !== 6) {
      toast.error("Please enter a valid 6-digit code");
      return;
    }
    setOtpLoading(true);
    router.push(url);
  };

  return (
    <Box className="max-w-xl px-6 md:w-full">
      <Logo asIcon className="h-10" />
      {accountDeleted ? (
        <Box
          className="border-success/20 bg-success/10 mt-6 rounded-lg border p-4"
          role="status"
        >
          <Text>Your account has been deleted.</Text>
          <Text>
            Shared workspace contributions remain attributed to Former user.
          </Text>
          {cleanupPending ? (
            <Text className="mt-1" color="muted">
              Connected-service cleanup will continue in the background.
            </Text>
          ) : null}
        </Box>
      ) : null}
      {isSent ? (
        <>
          <Text
            as="h1"
            className="mt-10 mb-2 text-3xl md:text-4xl"
            fontWeight="semibold"
          >
            Check your email
          </Text>
          <Text className="mb-6 pl-0.5" color="muted" fontWeight="medium">
            A secure sign-in {isMobileApp ? "code" : "link"} has been sent to{" "}
            <span className="font-semibold dark:text-white/70">{email}</span>.
            ✨ Please check your inbox to continue.
          </Text>
          <Box className="mb-4">
            <OTPInput className="mb-4" onChange={setOtp} value={otp} />
            <Button
              align="center"
              className="mb-4 md:py-3"
              color="invert"
              disabled={otp.length !== 6}
              fullWidth
              loading={otpLoading}
              loadingText="Verifying..."
              onClick={handleOTPSubmit}
              size="lg"
            >
              Verify Code
            </Button>
          </Box>
          <Text className="mb-6 pl-0.5" fontWeight="medium">
            Back to{" "}
            <button
              className="text-primary underline"
              onClick={() => {
                setIsSent(false);
              }}
              type="button"
            >
              Login
            </button>
          </Text>
        </>
      ) : (
        <>
          <Text
            as="h1"
            className="mt-10 mb-4 text-3xl md:text-4xl"
            fontWeight="semibold"
          >
            {page === "login"
              ? "Sign into your account"
              : "Create your account"}
          </Text>
          {page === "login" ? (
            <>
              {isMobileApp ? (
                <Text className="mb-6 pl-0.5" color="muted" fontWeight="medium">
                  Welcome back! sign in to your account to continue.
                </Text>
              ) : (
                <Text className="mb-6 pl-0.5" color="muted" fontWeight="medium">
                  Don&apos;t have an account?{" "}
                  <Link
                    className="text-primary underline"
                    href={withCallbackUrl("/signup", safeCallbackUrl)}
                  >
                    Create one
                  </Link>
                </Text>
              )}
            </>
          ) : (
            <Text className="mb-6 pl-0.5" color="muted" fontWeight="medium">
              Already have an account?{" "}
              <Link
                className="text-primary underline"
                href={withCallbackUrl(
                  isMobileApp ? "/?mobileApp=true" : "/",
                  safeCallbackUrl,
                )}
              >
                Sign in
              </Link>
            </Text>
          )}
          {signInError ? (
            <Box
              className="border-danger/20 bg-danger/10 mb-6 rounded-lg border p-4"
              role="alert"
            >
              <Text>{signInError}</Text>
              {errorMessage === "account_unavailable" ? (
                <Link
                  className="text-primary mt-2 inline-block underline"
                  href="https://fortyone.app/contact"
                >
                  Contact support
                </Link>
              ) : null}
            </Box>
          ) : null}
          <form onSubmit={handleSubmit}>
            <Input
              className="rounded-lg"
              hasError={
                Boolean(errorMessage) && !signInError && !email && !isTouched
              }
              helpText={!signInError && !isTouched ? errorMessage : undefined}
              label="Enter your email"
              name="email"
              onChange={(e) => {
                setEmail(e.target.value);
                setIsTouched(true);
              }}
              placeholder="e.g john@company.com"
              required
              type="email"
              value={email}
            />
            <Button
              align="center"
              className="mt-4 md:py-3"
              color="invert"
              fullWidth
              loading={loading}
              loadingText="Logging you in..."
              size="lg"
              type="submit"
            >
              Continue
            </Button>
            {!isMobileApp ? (
              <>
                <Flex align="center" className="my-4 gap-4" justify="between">
                  <Box className="bg-surface-muted h-px w-full" />
                  <Text className="text-[0.95rem] opacity-40">OR</Text>
                  <Box className="bg-surface-muted h-px w-full" />
                </Flex>
                <Button
                  align="center"
                  className="mb-3 md:py-2.5"
                  color="tertiary"
                  fullWidth
                  leftIcon={<GoogleIcon />}
                  onClick={async () => {
                    await signInWithGoogle(
                      withCallbackUrl("/auth-callback", safeCallbackUrl),
                    ).catch((error) => {
                      toast.error("Google sign-in failed", {
                        description:
                          error instanceof Error
                            ? error.message
                            : "Please try again.",
                      });
                    });
                  }}
                  size="lg"
                  type="button"
                >
                  Continue with Google
                </Button>
                <Button
                  align="center"
                  className="mb-3 md:py-2.5"
                  color="tertiary"
                  fullWidth
                  leftIcon={<MicrosoftIcon />}
                  onClick={async () => {
                    try {
                      await signInWithMicrosoft(
                        withCallbackUrl("/auth-callback", safeCallbackUrl),
                      );
                    } catch (error) {
                      toast.error("Microsoft sign-in failed", {
                        description:
                          error instanceof Error
                            ? error.message
                            : "Please try again.",
                      });
                    }
                  }}
                  size="lg"
                  type="button"
                >
                  Continue with Microsoft
                </Button>
              </>
            ) : null}
          </form>
          <Text className="mt-3 pl-px text-[90%]" color="muted">
            {COPYRIGHT_NOTICE}
          </Text>
        </>
      )}
    </Box>
  );
};
