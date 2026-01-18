"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Input, Text, Button, Flex } from "ui";
import Link from "next/link";
import { toast } from "sonner";
import { useRouter, useSearchParams } from "next/navigation";
import { Logo, GoogleIcon } from "@/components/ui";
import { OTPInput } from "@/components/ui/otp-input";
import { requestMagicEmail } from "@/lib/actions/request-magic-email";
import { signInWithGoogle } from "@/lib/actions/sign-in";

export const AuthLayout = ({ page }: { page: "login" | "signup" }) => {
  const [loading, setLoading] = useState(false);
  const [email, setEmail] = useState("");
  const [isSent, setIsSent] = useState(false);
  const [isTouched, setIsTouched] = useState(false);
  const [otp, setOtp] = useState("");
  const [otpLoading, setOtpLoading] = useState(false);
  const searchParams = useSearchParams();
  const error = searchParams?.get("error");
  const isMobileApp = searchParams?.get("mobileApp") === "true";
  const router = useRouter();

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    const result = await requestMagicEmail(email, isMobileApp);
    if (result?.error?.message) {
      toast.error("Failed to send magic link", {
        description: result.error.message,
      });
    } else {
      setIsSent(true);
    }
    setLoading(false);
  };

  const handleOTPSubmit = async () => {
    let url = `/verify/${email}/${otp}`;
    if (isMobileApp) {
      url += "?mobileApp=true";
    }

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
            <OTPInput value={otp} onChange={setOtp} className="mb-4" />
            <Button
              align="center"
              className="mb-4"
              color="invert"
              fullWidth
              loading={otpLoading}
              loadingText="Verifying..."
              onClick={handleOTPSubmit}
              disabled={otp.length !== 6}
            >
              Verify Code
            </Button>
          </Box>
          <Text className="mb-6 pl-0.5" fontWeight="medium">
            Back to{" "}
            <button
              type="button"
              className="text-primary underline dark:text-white"
              onClick={() => {
                setIsSent(false);
              }}
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
                    className="text-primary underline dark:text-white"
                    href="/signup"
                  >
                    Create one
                  </Link>
                </Text>
              )}
            </>
          ) : (
            <Text className="mb-6 pl-0.5" color="muted" fontWeight="medium">
              Already have an account?{" "}
              <Link className="text-primary underline dark:text-white" href="/">
                Sign in
              </Link>
            </Text>
          )}
          <form onSubmit={handleSubmit}>
            <Input
              autoFocus
              className="rounded-[0.6rem]"
              hasError={Boolean(error) && !email && !isTouched}
              helpText={error && !isTouched ? error : undefined}
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
              type="submit"
              size="lg"
            >
              Continue
            </Button>
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
                  isMobileApp
                    ? "/auth-callback?mobileApp=true"
                    : "/auth-callback",
                );
              }}
              type="button"
              size="lg"
            >
              Continue with Google
            </Button>
          </form>
          <Text className="mt-3 pl-px text-[90%]" color="muted">
            &copy; {new Date().getFullYear()} &bull; Product of Complexus LLC
            &bull; All Rights Reserved.
          </Text>
        </>
      )}
    </Box>
  );
};
