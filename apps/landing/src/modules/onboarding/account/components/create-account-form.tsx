"use client";

import { Input, Button } from "ui";
import type { FormEvent } from "react";
import { useState } from "react";

export const CreateAccountForm = () => {
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setIsLoading(true);
  };

  return (
    <form className="space-y-5" onSubmit={handleSubmit}>
      <Input
        className="rounded-lg"
        label="Full Name"
        name="name"
        placeholder="Enter your full name"
        required
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
