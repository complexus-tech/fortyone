"use client";

import { Box, Input, Select, Text, Button } from "ui";
import type { ChangeEvent, FormEvent } from "react";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "next-auth/react";
import { createWorkspaceAction } from "@/lib/actions/create-workspace";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const CreateWorkspaceForm = () => {
  const router = useRouter();
  const { data: session } = useSession();
  const [isLoading, setIsLoading] = useState(false);
  const [form, setForm] = useState({
    name: "",
    slug: "",
    teamSize: "2-10",
  });
  const workspaces = session?.workspaces || [];

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      setIsLoading(true);
      const workspace = await createWorkspaceAction(form);
      if (workspaces.length === 0) {
        router.push("/onboarding/welcome");
      } else if (domain.includes("localhost")) {
        window.location.href = `http://${workspace?.slug}.${domain}/my-work`;
      } else {
        window.location.href = `https://${workspace?.slug}.${domain}/my-work`;
      }
    } catch (error) {
      setIsLoading(false);
    }
  };

  return (
    <form className="space-y-5" onSubmit={handleSubmit}>
      <Input
        className="rounded-lg"
        label="Organization"
        name="name"
        onChange={handleChange}
        placeholder="Enter organization name"
        required
        value={form.name}
      />
      <Input
        className="rounded-lg"
        label="Workspace URL"
        name="slug"
        onChange={handleChange}
        placeholder="your-workspace"
        required
        value={form.slug}
      />
      <Box>
        <Text as="label" className="mb-2 block font-medium">
          How many people will use this workspace?
        </Text>
        <Select value={form.teamSize}>
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
