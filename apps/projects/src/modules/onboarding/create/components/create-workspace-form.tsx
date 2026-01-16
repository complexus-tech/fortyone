"use client";

import { Box, Input, Select, Text, Button, Flex } from "ui";
import type { ChangeEvent, FormEvent } from "react";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { CloseIcon } from "icons";
import { createWorkspaceAction } from "@/lib/actions/create-workspace";
import { useDebounce } from "@/hooks";
import { checkWorkspaceAvailability } from "@/lib/queries/check-workspace-availability";
import type { ApiResponse } from "@/types";
import { useWorkspaces } from "@/lib/hooks/workspaces";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const CreateWorkspaceForm = () => {
  const router = useRouter();
  const checkAvailability = useDebounce<string>(async (slugToCheck) => {
    if (slugToCheck.length > 3) {
      setIsAvailable(true); // Reset availability while checking
      try {
        const res = await checkWorkspaceAvailability(slugToCheck);
        setIsAvailable(res.data?.available || false);
      } catch (error) {
        setIsAvailable(false);
      }
    }
  }, 1000);

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setForm((prev) => {
      const updates = { ...prev, [name]: value };

      // If it's the org name and hasn't been blurred, update slug too
      if (name === "name" && !hasOrgBlurred) {
        updates.slug = formatSlug(value).toLowerCase();
      }

      // Check availability when slug changes
      if (name === "slug" || (name === "name" && !hasOrgBlurred)) {
        checkAvailability(updates.slug.toLowerCase());
      }

      return updates;
    });
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!form.name || !form.slug) {
      toast.warning("Validation warning", {
        description: "Please enter a workspace name and URL",
      });
      return;
    }

    if (form.slug.length > 16) {
      toast.warning("Validation warning", {
        description: "Workspace URL must be less than 16 characters",
      });
      return;
    }

    if (form.slug.length < 3) {
      toast.warning("Validation warning", {
        description: "Workspace URL must be at least 3 characters",
      });
      return;
    }

    setIsLoading(true);
    const res = await createWorkspaceAction({
      name: form.name.trim(),
      slug: form.slug.trim(),
      teamSize: form.teamSize,
    });
    if (res.error?.message) {
      setIsLoading(false);
      toast.error("Failed to create workspace", {
        description: res.error.message || "Something went wrong",
      });
      return;
    }
    const workspace = res.data!;
    if (workspaces.length === 0) {
      router.push("/onboarding/account");
    } else {
      window.location.href = `/${workspace?.slug}/my-work`;
    }
  };

  return (
    <form className="space-y-5" onSubmit={handleSubmit}>
      <Input
        className="rounded-[0.6rem]"
        label="Your Workspace"
        name="name"
        onBlur={() => {
          setHasOrgBlurred(true);
        }}
        onChange={handleChange}
        placeholder="Enter workspace name"
        required
        value={form.name}
      />
      <Input
        className="rounded-[0.6rem]"
        hasError={!isAvailable}
        helpText={
          !isAvailable
            ? "This URL is already taken. Please try a different one."
            : "Pick a simple, memorable URL for your workspace"
        }
        label="Workspace URL"
        maxLength={16}
        minLength={3}
        name="slug"
        onChange={handleChange}
        pattern="^[a-z][a-z0-9-]*$"
        required
        rightIcon={
          <Flex align="center" gap={2}>
            <Text>{`/${form.slug || "slug"}`}</Text>
            {!isAvailable ? (
              <Flex
                align="center"
                className="bg-danger size-5 rounded-full"
                justify="center"
              >
                <CloseIcon
                  className="h-3 text-white dark:text-white"
                  strokeWidth={3}
                />
              </Flex>
            ) : null}
          </Flex>
        }
        value={form.slug}
      />
      <Box>
        <Text as="label" className="mb-2 block font-medium">
          How many people will use this workspace?
        </Text>
        <Select
          onValueChange={(value) => {
            setForm({ ...form, teamSize: value });
          }}
          value={form.teamSize}
        >
          <Select.Trigger className="border-border bg-surface/70 h-[2.7rem] w-full rounded-[0.6rem]">
            <Select.Input />
          </Select.Trigger>
          <Select.Content>
            <Select.Group>
              <Select.Option
                className="h-10 rounded-[0.6rem] text-[0.9rem]"
                value="1-5"
              >
                1-5
              </Select.Option>
              <Select.Option
                className="h-10 rounded-[0.6rem] text-[0.9rem]"
                value="6-10"
              >
                6-10
              </Select.Option>
              <Select.Option
                className="h-10 rounded-[0.6rem] text-[0.9rem]"
                value="51-200"
              >
                51-200
              </Select.Option>
              <Select.Option
                className="h-10 rounded-[0.6rem] text-[0.9rem]"
                value="201-500"
              >
                201-500
              </Select.Option>
              <Select.Option
                className="h-10 rounded-[0.6rem] text-[0.9rem]"
                value="500+"
              >
                500+
              </Select.Option>
            </Select.Group>
          </Select.Content>
        </Select>
      </Box>
      <Button
        align="center"
        className="mt-4"
        color="invert"
        size="lg"
        fullWidth
        loading={isLoading}
        loadingText="Creating workspace..."
      >
        Create Workspace
      </Button>
    </form>
  );
};
