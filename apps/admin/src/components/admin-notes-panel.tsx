"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { PlusIcon } from "icons";
import { toast } from "sonner";
import { Box, Button, Flex, Text, TextArea } from "ui";
import { createAdminNote } from "@/lib/admin-api";
import { formatDateTime } from "@/lib/format";
import type { AdminNote } from "@/lib/types";

export const AdminNotesPanel = ({
  notes,
  targetId,
  targetType,
  workspaceId,
}: {
  notes: AdminNote[];
  targetId: string;
  targetType: "user" | "workspace";
  workspaceId?: string;
}) => {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [body, setBody] = useState("");

  const handleSubmit = (formData: FormData) => {
    const noteBody = String(formData.get("body") ?? "").trim();
    if (!noteBody) {
      toast.error("Note body is required.");
      return;
    }

    startTransition(async () => {
      try {
        await createAdminNote({
          targetType,
          targetId,
          workspaceId,
          body: noteBody,
        });
        setBody("");
        toast.success("Note added");
        router.refresh();
      } catch (error) {
        const message =
          error instanceof Error ? error.message : "Unable to add note.";
        toast.error("Note could not be added", { description: message });
      }
    });
  };

  return (
    <Box className="border-border bg-surface rounded-lg border-[0.5px]">
      <Box className="border-border border-b-[0.5px] px-4 py-3">
        <Text fontWeight="semibold">Admin notes</Text>
        <Text className="mt-1 text-[0.95rem]" color="muted">
          Internal support context attached to this {targetType}.
        </Text>
      </Box>
      <Box className="space-y-4 p-4">
        <form action={handleSubmit} className="space-y-3">
          <TextArea
            className="min-h-24 leading-6"
            name="body"
            onChange={(event) => {
              setBody(event.currentTarget.value);
            }}
            placeholder="Add operational context, support follow-up, billing notes..."
            value={body}
          />
          <Flex justify="end">
            <Button
              color="tertiary"
              disabled={isPending}
              loading={isPending}
              loadingText="Adding..."
              size="sm"
              type="submit"
            >
              <PlusIcon className="h-4" />
              Add note
            </Button>
          </Flex>
        </form>

        <Box className="divide-border divide-y">
          {notes.length > 0 ? (
            notes.map((note) => (
              <Box className="py-3 first:pt-0 last:pb-0" key={note.id}>
                <Text className="leading-6">{note.body}</Text>
                <Text className="mt-1 text-[0.95rem]" color="muted">
                  {note.createdByName || note.createdByEmail || "Unknown admin"}{" "}
                  · {formatDateTime(note.createdAt)}
                </Text>
              </Box>
            ))
          ) : (
            <Box className="py-6 text-center">
              <Text color="muted">No notes added yet.</Text>
            </Box>
          )}
        </Box>
      </Box>
    </Box>
  );
};
