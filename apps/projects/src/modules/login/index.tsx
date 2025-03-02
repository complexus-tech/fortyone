"use client";
import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Flex, Input, Text, Button } from "ui";
import { toast } from "sonner";
import nProgress from "nprogress";
import { redirect } from "next/navigation";
import { ComplexusLogo } from "@/components/ui";
import { logIn } from "./actions";

export const LoginPage = () => {
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    setLoading(true);
    nProgress.start();

    const res = await logIn(
      formData.get("email") as string,
      formData.get("password") as string,
    );

    if (res?.error) {
      toast.error("Failed to log in", {
        description: res.error,
      });
      setLoading(false);
      nProgress.done();
      return;
    }
    setLoading(false);
    nProgress.done();
    redirect("/my-work");
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
          Log in to your account
        </Text>
        <form onSubmit={handleSubmit}>
          <Input
            autoFocus
            className="mb-4 h-12 rounded-lg"
            label="Email"
            name="email"
            placeholder="e.g john@example.com"
            required
            type="email"
          />
          <Input
            className="mb-5 h-12 rounded-lg"
            label="Password"
            name="password"
            required
            type="password"
          />
          <Button
            align="center"
            className="md:h-12"
            fullWidth
            loading={loading}
            loadingText="Logging in..."
            type="submit"
          >
            Login
          </Button>
        </form>
        <Text className="textt-[90%] mt-3 pl-[1px]" color="muted">
          &copy; {new Date().getFullYear()} &bull; Powered by Complexus &bull;
          All Rights Reserved.
        </Text>
      </Box>
    </Flex>
  );
};
