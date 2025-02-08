"use client";

import { Box, Input, Select, Text, Button } from "ui";
import type { ChangeEvent } from "react";
import { useState } from "react";

export const CreateWorkspaceForm = () => {
  const [form, setForm] = useState({
    name: "",
    url: "",
    teamSize: "2-10",
  });

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
  };

  return (
    <Box className="space-y-5">
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
        name="url"
        onChange={handleChange}
        placeholder="your-workspace"
        required
        value={form.url}
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
        href="/onboarding/personalize"
        rounded="lg"
        size="lg"
      >
        Create Workspace
      </Button>
    </Box>
  );
};
