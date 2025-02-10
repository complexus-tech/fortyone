"use client";

import { Box, Input, Select, Text, Button } from "ui";
import type { ChangeEvent, FormEvent } from "react";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "next-auth/react";
import { createWorkspaceAction } from "@/lib/actions/workspaces/create-workspace";

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
      await createWorkspaceAction(form);
      if (workspaces.length === 0) {
        localStorage.clear();
        router.push("/onboarding/personalize");
      } else {
        window.location.href = "/my-work";
      }
    } catch (error) {
      setIsLoading(false);
    }
  };

  return (
    <form className="space-y-5" onSubmit={handleSubmit}>
      <Input
        className="rounded-lg md:h-[3rem] md:leading-[3rem]"
        label="Workspace Name"
        name="name"
        onChange={handleChange}
        placeholder="Enter workspace name"
        required
        value={form.name}
      />
      <Input
        className="rounded-lg md:h-[3rem] md:leading-[3rem]"
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
          <Select.Trigger className="h-[3rem] w-full rounded-lg border-gray-100 bg-white/70 text-base dark:border-dark-100 dark:bg-dark/20">
            <Select.Input />
          </Select.Trigger>
          <Select.Content>
            <Select.Group>
              <Select.Option className="h-[2.6rem] text-base" value="2-10">
                2-10
              </Select.Option>
              <Select.Option className="h-[2.6rem] text-base" value="11-50">
                11-50
              </Select.Option>
              <Select.Option className="h-[2.6rem] text-base" value="51-200">
                51-200
              </Select.Option>
              <Select.Option className="h-[2.6rem] text-base" value="201-500">
                201-500
              </Select.Option>
              <Select.Option className="h-[2.6rem] text-base" value="500+">
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
        rounded="lg"
        size="lg"
      >
        Create Workspace
      </Button>
    </form>
  );
};
