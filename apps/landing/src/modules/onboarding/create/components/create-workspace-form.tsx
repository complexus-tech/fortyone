"use client";

import { Box, Input, Select, Text, Button, Flex } from "ui";
import type { ChangeEvent, FormEvent } from "react";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "next-auth/react";
import { toast } from "sonner";
import { CloseIcon } from "icons";
import { createWorkspaceAction } from "@/lib/actions/create-workspace";
import { useDebounce } from "@/hooks";
import { checkWorkspaceAvailability } from "@/lib/actions/check-workspace-availability";
import type { ApiResponse } from "@/types";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const CreateWorkspaceForm = () => {
  const router = useRouter();
  const { data: session } = useSession();
  const checkAvailability = useDebounce<
    string,
    ApiResponse<{
      available: boolean;
      slug: string;
    }>
  >(checkWorkspaceAvailability, 1000);
  const [isLoading, setIsLoading] = useState(false);
  const [isAvailable, setIsAvailable] = useState(true);
  const [hasOrgBlurred, setHasOrgBlurred] = useState(false);
  const [form, setForm] = useState({
    name: "",
    slug: "",
    teamSize: "2-10",
  });
  const prevWorkspaces = session?.workspaces || [];

  const formatSlug = (name: string) => {
    return name
      .toLowerCase()
      .replace(/[^a-z0-9-]/g, "-") // Replace non-alphanumeric chars with dash
      .replace(/-+/g, "-") // Replace multiple dashes with single dash
      .replace(/^-|-$/g, ""); // Remove leading/trailing dashes
  };

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
        const slugToCheck = updates.slug.toLowerCase();
        if (slugToCheck.length > 3) {
          setIsAvailable(true); // Reset availability while checking
          void checkAvailability(slugToCheck)
            .then((res) => {
              setIsAvailable(res.data?.available || false);
            })
            .catch(() => {
              setIsAvailable(false);
            });
        }
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
    const res = await createWorkspaceAction(form);
    if (res.error?.message) {
      setIsLoading(false);
      toast.error("Failed to create workspace", {
        description: res.error.message || "Something went wrong",
      });
      return;
    }
    const workspace = res.data!;

    if (prevWorkspaces.length === 0) {
      router.push("/onboarding/welcome");
    } else if (domain.includes("localhost")) {
      window.location.href = `http://${workspace?.slug}.${domain}/my-work`;
    } else {
      window.location.href = `https://${workspace?.slug}.${domain}/my-work`;
    }
  };

  return (
    <form className="space-y-5" onSubmit={handleSubmit}>
      <Input
        className="rounded-lg"
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
        className="rounded-lg"
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
          form.slug ? (
            <Flex align="center" gap={2}>
              <Text
                className="antialiased"
                color="muted"
                fontWeight="semibold"
              >{`.${domain}`}</Text>
              {!isAvailable ? (
                <Flex
                  align="center"
                  className="size-5 rounded-full bg-danger"
                  justify="center"
                >
                  <CloseIcon
                    className="h-3 text-white dark:text-white"
                    strokeWidth={3}
                  />
                </Flex>
              ) : null}
            </Flex>
          ) : null
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
          <Select.Trigger className="h-[2.7rem] w-full rounded-lg border-gray-100 bg-white/70 text-base dark:border-dark-100 dark:bg-dark/20">
            <Select.Input />
          </Select.Trigger>
          <Select.Content>
            <Select.Group>
              <Select.Option
                className="h-[2.5rem] rounded-lg text-[0.9rem]"
                value="2-10"
              >
                2-10
              </Select.Option>
              <Select.Option
                className="h-[2.5rem] rounded-lg text-[0.9rem]"
                value="11-50"
              >
                11-50
              </Select.Option>
              <Select.Option
                className="h-[2.5rem] rounded-lg text-[0.9rem]"
                value="51-200"
              >
                51-200
              </Select.Option>
              <Select.Option
                className="h-[2.5rem] rounded-lg text-[0.9rem]"
                value="201-500"
              >
                201-500
              </Select.Option>
              <Select.Option
                className="h-[2.5rem] rounded-lg text-[0.9rem]"
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
        fullWidth
        loading={isLoading}
        loadingText="Creating workspace..."
      >
        Create Workspace
      </Button>
    </form>
  );
};
