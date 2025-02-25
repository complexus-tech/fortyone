"use client";

import { Input, Button } from "ui";
import type { FormEvent } from "react";
import { useState } from "react";
import { toast } from "sonner";
import { updateProfile } from "@/lib/actions/update-profile";

export const CreateAccountForm = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [fullName, setFullName] = useState("");

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!fullName) {
      toast.warning("Validation error", {
        description: "Full name is required",
      });
      return;
    }

    setIsLoading(true);

    const _ = await updateProfile({ fullName });

    setIsLoading(false);
  };

  return (
    <form className="w-full space-y-5" onSubmit={handleSubmit}>
      <Input
        className="rounded-lg"
        label="Full Name"
        minLength={3}
        name="fullName"
        onChange={(e) => {
          setFullName(e.target.value);
        }}
        placeholder="Enter your full name"
        required
        value={fullName}
      />
      <Button
        align="center"
        className="mt-4"
        fullWidth
        loading={isLoading}
        loadingText="Creating account..."
      >
        Create Account
      </Button>
    </form>
  );
};
