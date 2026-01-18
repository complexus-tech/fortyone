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
  const currentTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
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
      await updateProfile({ fullName, timezone: currentTimezone });
    } catch (error) {
      setIsLoading(false);
    } finally {
      router.push("/onboarding/invite");
    }
  };

  return (
    <form className="w-full space-y-4" onSubmit={handleSubmit}>
      <Input
        className="rounded-[0.6rem]"
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
        className="flex items-start gap-2 select-none"
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
          FortyOne.
        </Text>
      </label>

      <Button
        align="center"
        className="mt-4 md:py-3"
        color="invert"
        fullWidth
        size="lg"
        loading={isLoading}
        loadingText="Creating account..."
      >
        Create Account
      </Button>
    </form>
  );
};
