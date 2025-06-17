"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Input, Text, Button, Flex } from "ui";
import Link from "next/link";
import { toast } from "sonner";
import { useSearchParams } from "next/navigation";
import { Logo, GoogleIcon } from "@/components/ui";
import { requestMagicEmail } from "@/lib/actions/request-magic-email";
import { signInWithGoogle } from "@/lib/actions/sign-in";

export const AuthLayout = ({ page }: { page: "login" | "signup" }) => {
  const [loading, setLoading] = useState(false);
  const [email, setEmail] = useState("");
  const [isSent, setIsSent] = useState(false);
  const [isTouched, setIsTouched] = useState(false);
  const searchParams = useSearchParams();
  const error = searchParams?.get("error");

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    const result = await requestMagicEmail(email);
    if (result?.error?.message) {
      toast.error("Failed to send magic link", {
        description: result.error.message,
      });
      setLoading(false);
    } else {
      setIsSent(true);
    }
  };

  return (
    <Box className="max-w-sm">
      <Logo asIcon className="relative -left-2 h-10 text-white" />
      {isSent ? (
        <>
          <Text
            as="h1"
            className="mb-2 mt-6 text-[1.7rem]"
            fontWeight="semibold"
          >
            Check your email
          </Text>
          <Text className="mb-6 pl-0.5" color="muted" fontWeight="medium">
            A secure sign-in link has been sent to{" "}
            <span className="font-semibold text-white/70">{email}</span>. ✨
            Please check your inbox to continue.
          </Text>
        </>
      ) : (
        <>
          <Text
            as="h1"
            className="mb-2 mt-6 text-[1.7rem]"
            fontWeight="semibold"
          >
            {page === "login"
              ? "Sign in to Complexus"
              : "Get started with Complexus"}
          </Text>
          {page === "login" ? (
            <Text className="mb-6 pl-0.5" color="muted" fontWeight="medium">
              Don&apos;t have an account?{" "}
              <Link className="text-primary" href="/signup">
                Create one
              </Link>
            </Text>
          ) : (
            <Text className="mb-6 pl-0.5" color="muted" fontWeight="medium">
              Already have an account?{" "}
              <Link className="text-primary" href="/login">
                Sign in
              </Link>
            </Text>
          )}

          <form onSubmit={handleSubmit}>
            <Input
              autoFocus
              className="rounded-lg"
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
              className="mt-4"
              fullWidth
              loading={loading}
              loadingText="Logging you in..."
              type="submit"
            >
              Continue
            </Button>
            <Flex align="center" className="my-4 gap-4" justify="between">
              <Box className="h-px w-full bg-gray-100 dark:bg-white/20" />
              <Text className="text-[0.95rem] opacity-40">OR</Text>
              <Box className="h-px w-full bg-gray-100 dark:bg-white/20" />
            </Flex>
            <Button
              align="center"
              className="mb-3 border-gray-200 md:h-[2.6rem]"
              color="tertiary"
              fullWidth
              leftIcon={<GoogleIcon />}
              onClick={async () => {
                await signInWithGoogle();
              }}
              type="button"
            >
              Continue with Google
            </Button>
          </form>
          <Text className="mt-3 pl-[1px] text-[90%]" color="muted">
            &copy; {new Date().getFullYear()} &bull; Powered by Complexus &bull;
            All Rights Reserved.
          </Text>
        </>
      )}
    </Box>
  );
};
