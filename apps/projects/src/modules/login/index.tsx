"use client";
import { FormEvent, useState } from "react";
import { Box, Flex, Input, Text, Button } from "ui";
import { ComplexusLogo } from "@/components/ui";
import { logIn } from "./actions";
import { toast } from "sonner";
import nProgress from "nprogress";

export const LoginPage = ({ callbackUrl }: { callbackUrl: string }) => {
  const logInAction = logIn.bind(null, callbackUrl);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    setLoading(true);
    nProgress.start();
    try {
      const result = await logInAction(formData);
      if (result?.error) {
        toast.error("Failed to log in", {
          description: result.error,
        });
      }
    } catch (_) {
    } finally {
      setLoading(false);
      nProgress.done();
    }
  };

  return (
    <Flex align="center" justify="center" className="h-screen">
      <Box className="mx-auto w-full max-w-md rounded-xl">
        <Box className="mx-auto w-max rounded-full bg-primary p-3.5">
          <ComplexusLogo className="h-8 w-auto text-white" />
        </Box>
        <Text
          as="h1"
          fontSize="3xl"
          fontWeight="medium"
          align="center"
          className="mb-10 mt-3"
        >
          Log in to your account
        </Text>
        <form onSubmit={handleSubmit}>
          <Input
            label="Email"
            type="email"
            autoFocus
            required
            name="email"
            placeholder="e.g john@example.com"
            className="mb-4 h-12 rounded-lg"
          />
          <Input
            label="Password"
            required
            name="password"
            type="password"
            className="mb-5 h-12 rounded-lg"
          />
          <Button
            loadingText="Logging in..."
            type="submit"
            loading={loading}
            className="md:h-12"
            align="center"
            fullWidth
          >
            Login
          </Button>
        </form>
        <Text color="muted" className="textt-[90%] mt-3 pl-[1px]">
          &copy; {new Date().getFullYear()} &bull; Powered by Complexus &bull;
          All Rights Reserved.
        </Text>
      </Box>
    </Flex>
  );
};
