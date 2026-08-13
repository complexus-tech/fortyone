"use client";

import { useMemo, useState } from "react";
import { DeleteIcon, ExternalLinkIcon, PlusIcon, SearchIcon } from "icons";
import { Box, Button, Dialog, Flex, Input, Text } from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useDebouncedCallback } from "@/hooks/debounce";
import { SectionHeader } from "@/modules/settings/components";
import type {
  FeedbackItemCandidate,
  FeedbackPortal,
  FeedbackUpdate,
} from "./types";
import {
  useCreateFeedbackUpdateMutation,
  useDeleteFeedbackUpdateMutation,
  useFeedbackUpdateCandidates,
  useFeedbackUpdates,
  usePublishFeedbackUpdateMutation,
  useUpdateFeedbackUpdateMutation,
} from "./hooks";

type EditorTarget = FeedbackUpdate | "new";

const UpdateEditor = ({
  onClose,
  portal,
  target,
}: {
  onClose: () => void;
  portal: FeedbackPortal;
  target: EditorTarget;
}) => {
  const current = target === "new" ? null : target;
  const [title, setTitle] = useState(current?.title ?? "");
  const [summary, setSummary] = useState(current?.summary ?? "");
  const [body, setBody] = useState(current?.body ?? "");
  const [coverImageUrl, setCoverImageUrl] = useState(
    current?.coverImageUrl ?? "",
  );
  const [itemIds, setItemIds] = useState(
    () => new Set(current?.linkedItems.map((item) => item.id) ?? []),
  );
  const [search, setSearch] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const { callback: searchCandidates } = useDebouncedCallback<string>(
    setSearchQuery,
    300,
  );
  const candidatesQuery = useFeedbackUpdateCandidates(portal.id, searchQuery);
  const createMutation = useCreateFeedbackUpdateMutation();
  const updateMutation = useUpdateFeedbackUpdateMutation();
  const isPending = createMutation.isPending || updateMutation.isPending;
  const visibleCandidates = useMemo(() => {
    const linkedItems: FeedbackItemCandidate[] = (
      current?.linkedItems ?? []
    ).map((item) => ({ ...item, commentCount: 0, voteCount: 0 }));
    const candidates = candidatesQuery.data?.candidates ?? [];
    const items = searchQuery ? candidates : [...linkedItems, ...candidates];
    return [...new Map(items.map((item) => [item.id, item])).values()];
  }, [candidatesQuery.data?.candidates, current?.linkedItems, searchQuery]);

  const save = async () => {
    const input = {
      body: body.trim(),
      coverImageUrl: coverImageUrl.trim() || undefined,
      itemIds: [...itemIds],
      portalId: portal.id,
      summary: summary.trim() || undefined,
      title: title.trim(),
    };
    if (current) {
      await updateMutation.mutateAsync({ input, updateId: current.id });
    } else {
      await createMutation.mutateAsync(input);
    }
    onClose();
  };

  return (
    <Dialog.Content className="max-w-3xl">
      <Dialog.Header>
        <Dialog.Title className="px-6 pt-1 text-lg">
          {current ? "Edit update" : "Create update"}
        </Dialog.Title>
      </Dialog.Header>
      <Dialog.Body className="max-h-[75dvh] space-y-5 overflow-y-auto">
        <label className="block space-y-2" htmlFor="feedback-update-title">
          <Text as="span" className="block text-sm" fontWeight="medium">
            Title
          </Text>
          <Input
            autoFocus
            id="feedback-update-title"
            maxLength={200}
            onChange={(event) => {
              setTitle(event.target.value);
            }}
            placeholder="What changed?"
            value={title}
          />
        </label>
        <label className="block space-y-2" htmlFor="feedback-update-summary">
          <Text as="span" className="block text-sm" fontWeight="medium">
            Summary
          </Text>
          <Input
            id="feedback-update-summary"
            maxLength={300}
            onChange={(event) => {
              setSummary(event.target.value);
            }}
            placeholder="A short description shown in update lists"
            value={summary}
          />
        </label>
        <label className="block space-y-2" htmlFor="feedback-update-body">
          <Text as="span" className="block text-sm" fontWeight="medium">
            Update
          </Text>
          <textarea
            className="border-border bg-surface ring-ring min-h-48 w-full resize-y rounded-xl border p-3 text-sm leading-6 outline-none placeholder:text-[var(--color-text-muted)] focus-visible:ring-2"
            id="feedback-update-body"
            maxLength={20000}
            onChange={(event) => {
              setBody(event.target.value);
            }}
            placeholder="Explain what shipped, why it matters, and what happens next."
            value={body}
          />
        </label>
        {body.trim() ? (
          <Box className="border-border/70 bg-background rounded-xl border p-4">
            <Text
              className="mb-2 text-xs tracking-wide uppercase"
              color="muted"
            >
              Preview
            </Text>
            <Text className="text-sm leading-6 whitespace-pre-wrap">
              {body.trim()}
            </Text>
          </Box>
        ) : null}
        <label className="block space-y-2" htmlFor="feedback-update-cover">
          <Text as="span" className="block text-sm" fontWeight="medium">
            Cover image URL{" "}
            <span className="text-text-muted font-normal">(optional)</span>
          </Text>
          <Input
            id="feedback-update-cover"
            onChange={(event) => {
              setCoverImageUrl(event.target.value);
            }}
            placeholder="https://cdn.example.com/update.png"
            type="url"
            value={coverImageUrl}
          />
        </label>
        <Box>
          <Text className="text-sm" fontWeight="medium">
            Linked feedback
          </Text>
          <Text className="mt-1 text-sm" color="muted">
            Followers of linked requests are notified when this update is
            published.
          </Text>
          <Input
            className="mt-3"
            leftIcon={<SearchIcon className="h-4" />}
            onChange={(event) => {
              setSearch(event.target.value);
              searchCandidates(event.target.value);
            }}
            placeholder="Search feedback"
            value={search}
          />
          <Box className="border-border/70 mt-2 max-h-52 overflow-y-auto rounded-xl border">
            {candidatesQuery.isLoading ? (
              <Text className="px-4 py-8 text-center text-sm" color="muted">
                Loading feedback…
              </Text>
            ) : null}
            {!candidatesQuery.isLoading && visibleCandidates.length > 0
              ? visibleCandidates.map((item) => (
                  <Flex
                    align="start"
                    className="border-border/60 hover:bg-state-hover/40 flex cursor-pointer items-start gap-3 border-b px-3 py-3 last:border-b-0"
                    key={item.id}
                  >
                    <input
                      checked={itemIds.has(item.id)}
                      className="mt-0.5 size-4"
                      id={`feedback-update-item-${item.id}`}
                      onChange={(event) => {
                        setItemIds((current) => {
                          const next = new Set(current);
                          if (event.target.checked) next.add(item.id);
                          else next.delete(item.id);
                          return next;
                        });
                      }}
                      type="checkbox"
                    />
                    <label
                      className="min-w-0 flex-1 cursor-pointer"
                      htmlFor={`feedback-update-item-${item.id}`}
                    >
                      <Text className="truncate text-sm" fontWeight="medium">
                        {item.title}
                      </Text>
                      <Text className="mt-0.5 text-xs capitalize" color="muted">
                        {item.status.replace("_", " ")}
                      </Text>
                    </label>
                  </Flex>
                ))
              : null}
            {!candidatesQuery.isLoading && visibleCandidates.length === 0 ? (
              <Text className="px-4 py-8 text-center text-sm" color="muted">
                No matching feedback
              </Text>
            ) : null}
          </Box>
        </Box>
      </Dialog.Body>
      <Dialog.Footer className="justify-end gap-3">
        <Button color="tertiary" onClick={onClose}>
          Cancel
        </Button>
        <Button
          color="primary"
          disabled={!title.trim() || !body.trim() || isPending}
          loading={isPending}
          onClick={() => {
            void save();
          }}
        >
          Save draft
        </Button>
      </Dialog.Footer>
    </Dialog.Content>
  );
};

export const FeedbackUpdatesSettings = ({
  portal,
}: {
  portal: FeedbackPortal;
}) => {
  const updatesQuery = useFeedbackUpdates();
  const publishMutation = usePublishFeedbackUpdateMutation();
  const deleteMutation = useDeleteFeedbackUpdateMutation();
  const [editorTarget, setEditorTarget] = useState<EditorTarget | null>(null);
  const [publishTarget, setPublishTarget] = useState<FeedbackUpdate | null>(
    null,
  );
  const [deleteTarget, setDeleteTarget] = useState<FeedbackUpdate | null>(null);
  const updates = updatesQuery.data ?? [];

  return (
    <Box className="bg-surface overflow-hidden rounded-2xl">
      <SectionHeader
        action={
          <Button
            color="tertiary"
            leftIcon={<PlusIcon className="h-4" />}
            onClick={() => {
              setEditorTarget("new");
            }}
            size="sm"
          >
            New update
          </Button>
        }
        description="Publish product news, connect it to shipped feedback, and notify followers."
        title="Updates"
      />
      {updatesQuery.isLoading ? (
        <Text className="px-6 py-8" color="muted">
          Loading updates…
        </Text>
      ) : null}
      {!updatesQuery.isLoading && updates.length === 0 ? (
        <Box className="px-6 py-10 text-center">
          <UpdatesEmpty />
        </Box>
      ) : null}
      {!updatesQuery.isLoading && updates.length > 0 ? (
        <Box>
          {updates.map((update) => (
            <Flex
              align="center"
              className="border-border/60 flex-col items-stretch gap-3 border-b px-6 py-4 last:border-b-0 sm:flex-row sm:items-center"
              justify="between"
              key={update.id}
            >
              <Box className="min-w-0">
                <Flex align="center" className="gap-2">
                  <Text className="truncate" fontWeight="medium">
                    {update.title}
                  </Text>
                  <span
                    className={`rounded-md px-2 py-0.5 text-[11px] font-medium ${
                      update.publishedAt
                        ? "bg-emerald-500/10 text-emerald-700 dark:text-emerald-400"
                        : "bg-surface-muted text-text-muted"
                    }`}
                  >
                    {update.publishedAt ? "Published" : "Draft"}
                  </span>
                </Flex>
                <Text className="mt-1 truncate text-sm" color="muted">
                  {update.summary ||
                    `${update.linkedItems.length} linked feedback request${update.linkedItems.length === 1 ? "" : "s"}`}
                </Text>
              </Box>
              <Flex align="center" className="shrink-0 gap-2">
                {update.publishedAt ? (
                  <Button
                    aria-label={`Open ${update.title}`}
                    asIcon
                    color="tertiary"
                    href={`/portal/${portal.slug}/updates/${update.slug}`}
                    size="sm"
                    target="_blank"
                    variant="naked"
                  >
                    <ExternalLinkIcon className="h-4" />
                  </Button>
                ) : null}
                {!update.publishedAt ? (
                  <Button
                    color="tertiary"
                    onClick={() => {
                      setEditorTarget(update);
                    }}
                    size="sm"
                  >
                    Edit
                  </Button>
                ) : null}
                <Button
                  color="tertiary"
                  onClick={() => {
                    if (update.publishedAt) {
                      publishMutation.mutate({
                        publish: false,
                        updateId: update.id,
                      });
                    } else {
                      setPublishTarget(update);
                    }
                  }}
                  size="sm"
                >
                  {update.publishedAt ? "Unpublish" : "Publish"}
                </Button>
                {!update.publishedAt ? (
                  <Button
                    aria-label={`Delete ${update.title}`}
                    asIcon
                    color="tertiary"
                    onClick={() => {
                      setDeleteTarget(update);
                    }}
                    size="sm"
                    variant="naked"
                  >
                    <DeleteIcon className="h-4" />
                  </Button>
                ) : null}
              </Flex>
            </Flex>
          ))}
        </Box>
      ) : null}

      <Dialog
        onOpenChange={(open) => {
          if (!open) setEditorTarget(null);
        }}
        open={Boolean(editorTarget)}
      >
        {editorTarget ? (
          <UpdateEditor
            key={editorTarget === "new" ? "new" : editorTarget.id}
            onClose={() => {
              setEditorTarget(null);
            }}
            portal={portal}
            target={editorTarget}
          />
        ) : null}
      </Dialog>

      <ConfirmDialog
        confirmText="Publish update"
        description="Publishing makes this update public and emails eligible followers of every linked feedback request. Internal drafts are never shown publicly."
        isLoading={publishMutation.isPending}
        isOpen={Boolean(publishTarget)}
        loadingText="Publishing…"
        onClose={() => {
          setPublishTarget(null);
        }}
        onConfirm={() => {
          if (!publishTarget) return;
          publishMutation.mutate(
            { publish: true, updateId: publishTarget.id },
            {
              onSuccess: () => {
                setPublishTarget(null);
              },
            },
          );
        }}
        title="Publish this update?"
      />
      <ConfirmDialog
        confirmPhrase="delete"
        confirmText="Delete update"
        description="This permanently deletes the update. Published links to it will stop working."
        isLoading={deleteMutation.isPending}
        isOpen={Boolean(deleteTarget)}
        loadingText="Deleting…"
        onClose={() => {
          setDeleteTarget(null);
        }}
        onConfirm={() => {
          if (!deleteTarget) return;
          deleteMutation.mutate(deleteTarget.id, {
            onSuccess: () => {
              setDeleteTarget(null);
            },
          });
        }}
        title="Delete this update?"
      />
    </Box>
  );
};

const UpdatesEmpty = () => (
  <Box>
    <Text className="text-base" fontWeight="semibold">
      No updates yet
    </Text>
    <Text className="mx-auto mt-1 max-w-lg text-sm leading-6" color="muted">
      Create a draft when you have progress to share. Updates stay private until
      you explicitly publish them.
    </Text>
  </Box>
);
