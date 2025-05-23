"use client";

import { Input, Button, Checkbox, Text } from "ui";
import type { FormEvent } from "react";
import { useState } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useSession } from "next-auth/react";
import { updateProfile } from "@/lib/actions/update-profile";

export const CreateAccountForm = () => {
  const { data: session } = useSession();
  const [isLoading, setIsLoading] = useState(false);
  const [fullName, setFullName] = useState(session?.user?.name || "");
  const [receiveUpdates, setReceiveUpdates] = useState(false);
  const router = useRouter();

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!fullName) {
      toast.warning("Validation error", {
        description: "Full name is required",
      });
      return;
    }
    try {
      setIsLoading(true);
      const _ = await updateProfile({ fullName });
    } catch (error) {
      setIsLoading(false);
    } finally {
      router.push("/onboarding/invite");
    }
  };

  return (
    <form className="w-full space-y-4" onSubmit={handleSubmit}>
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
      <label
        className="flex select-none items-start gap-2"
        htmlFor="receive-updates"
      >
        <Checkbox
          checked={receiveUpdates}
          className="relative top-1 shrink-0"
          id="receive-updates"
          onCheckedChange={(checked) => {
            setReceiveUpdates(checked === "indeterminate" ? false : checked);
          }}
        />
        <Text color="muted">
          I&apos;d like to receive occasional product updates and tips from
          Complexus.
        </Text>
      </label>

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
