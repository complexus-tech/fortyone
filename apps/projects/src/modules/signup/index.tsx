"use client";
import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Flex, Input, Text, Button } from "ui";
import { toast } from "sonner";
import nProgress from "nprogress";
import Link from "next/link";
import { ComplexusLogo } from "@/components/ui";
import { signUp } from "@/lib/actions/users/sign-up";

export const SignUpPage = () => {
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const password = formData.get("password") as string;
    const confirmPassword = formData.get("confirmPassword") as string;

    if (password !== confirmPassword) {
      toast.error("Password validation failed", {
        description: "Passwords do not match",
      });
      return;
    }

    if (password.length < 8) {
      toast.error("Password validation failed", {
        description: "Password must be at least 8 characters",
      });
      return;
    }

    setLoading(true);
    nProgress.start();
    const result = await signUp(formData);
    if (result.error) {
      toast.error("Failed to create account", {
        description: result.error,
      });
      setLoading(false);
      nProgress.done();
    } else {
      toast.success("Account created successfully");
    }
  };

  return (
    <Flex align="center" className="h-screen" justify="center">
      <Box className="mx-auto w-full max-w-md rounded-xl">
        <Box className="mx-auto w-max rounded-full bg-primary p-3.5">
          <ComplexusLogo className="h-8 w-auto text-white" />
        </Box>
        <Text
          align="center"
          as="h1"
          className="mb-10 mt-3"
          fontSize="3xl"
          fontWeight="medium"
        >
          Create your account
        </Text>
        <form onSubmit={handleSubmit}>
          <Input
            autoFocus
            className="mb-4 h-12 rounded-lg"
            label="Full Name"
            name="fullName"
            placeholder="e.g John Doe"
            required
            type="text"
          />
          <Input
            className="mb-4 h-12 rounded-lg"
            label="Email"
            name="email"
            placeholder="e.g john@example.com"
            required
            type="email"
          />
          <Input
            className="mb-4 h-12 rounded-lg"
            label="Password"
            minLength={8}
            name="password"
            required
            type="password"
          />
          <Input
            className="mb-5 h-12 rounded-lg"
            label="Confirm Password"
            minLength={8}
            name="confirmPassword"
            required
            type="password"
          />
          <Button
            align="center"
            className="md:h-12"
            fullWidth
            loading={loading}
            loadingText="Creating account..."
            type="submit"
          >
            Create Account
          </Button>
        </form>
        <Link className="mt-4 block" href="/login">
          <Text
            align="center"
            className="underline underline-offset-1 hover:text-primary dark:hover:text-primary"
            color="muted"
          >
            Already have an account?
          </Text>
        </Link>
        <Text align="center" className="mt-3 pl-[1px] text-[90%]" color="muted">
          &copy; {new Date().getFullYear()} &bull; Powered by Complexus &bull;
          All Rights Reserved.
        </Text>
      </Box>
    </Flex>
  );
};
