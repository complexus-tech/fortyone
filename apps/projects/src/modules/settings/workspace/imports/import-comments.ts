import { z } from "zod";
import type { WorkspaceCtx } from "@/lib/http";
import { post } from "@/lib/http";
import type { ImportSourceComment } from "./schema";
import type { ImportedStories } from "./import-stories";
import { getBoundedImportSourceKey } from "./execution";

type CommentItem = ImportSourceComment & { storyId: string };
const commentResultSchema = z.object({
  sourceId: z.string(),
  storyId: z.uuid(),
  commentId: z.uuid().nullable(),
  error: z.object({ code: z.string(), message: z.string() }).nullable(),
});
const commentResponseSchema = z.object({
  data: z.object({ items: z.array(commentResultSchema) }).nullish(),
  error: z.object({ message: z.string() }).nullish(),
});
export type ImportCommentOutcome = {
  comments: number;
  unresolvedComments: number;
  commentIssues: string[];
};

// Long HTML is preserved as escaped source text instead of cutting through
// tags. Conservative text chunks leave room for attribution and escaping.
export const splitSourceComment = (
  comment: ImportSourceComment,
): ImportSourceComment[] => {
  const preserveHtmlAsText =
    comment.format === "html" && comment.content.length > 8500;
  const maximum =
    comment.format === "html" && !preserveHtmlAsText ? 8500 : 1000;
  if (comment.content.length <= maximum) return [comment];
  const characters = Array.from(comment.content);
  const chunks: ImportSourceComment[] = [];
  for (let index = 0; index < characters.length; index += maximum) {
    const part = chunks.length;
    chunks.push({
      ...comment,
      ...(preserveHtmlAsText ? { format: "text" as const } : {}),
      sourceId: part
        ? `${comment.sourceId}:part:${part + 1}`
        : comment.sourceId,
      content: characters.slice(index, index + maximum).join(""),
      parentSourceId: part ? comment.sourceId : comment.parentSourceId,
    });
  }
  return chunks;
};

const orderComments = (comments: ImportSourceComment[]) => {
  const byId = new Map(comments.map((comment) => [comment.sourceId, comment]));
  if (byId.size !== comments.length)
    throw new Error("Source comments contain duplicate IDs.");
  const ordered: ImportSourceComment[] = [];
  const completed = new Set<string>();
  const visiting = new Set<string>();
  const visit = (comment: ImportSourceComment) => {
    if (completed.has(comment.sourceId)) return;
    if (visiting.has(comment.sourceId))
      throw new Error("Source comment replies contain a cycle.");
    visiting.add(comment.sourceId);
    if (comment.parentSourceId) {
      const parent = byId.get(comment.parentSourceId);
      if (parent) visit(parent);
      else throw new Error("A source comment reply has a missing parent.");
    }
    visiting.delete(comment.sourceId);
    completed.add(comment.sourceId);
    ordered.push(comment);
  };
  comments.forEach(visit);
  return ordered;
};
export const importComments = async (
  {
    ctx,
    onProgress,
  }: { ctx: WorkspaceCtx; onProgress: (value: number) => void },
  {
    preparedTasks,
    importedStoryIdsBySourceKey,
  }: Pick<ImportedStories, "preparedTasks" | "importedStoryIdsBySourceKey">,
): Promise<ImportCommentOutcome> => {
  const items: CommentItem[] = [];
  let unresolvedComments = 0;
  const commentIssues: string[] = [];
  for (const item of preparedTasks) {
    const sourceComments = item.task.comments ?? [];
    if (!sourceComments.length) continue;
    const storyId = importedStoryIdsBySourceKey.get(item.sourceKey);
    if (!storyId) {
      unresolvedComments += sourceComments.length;
      continue;
    }
    try {
      items.push(
        ...orderComments(sourceComments).flatMap((comment) =>
          splitSourceComment(comment).map((part) => ({ ...part, storyId })),
        ),
      );
    } catch (error) {
      unresolvedComments += sourceComments.length;
      commentIssues.push(
        `${item.sourceId}: ${error instanceof Error ? error.message : "comments could not be mapped."}`,
      );
    }
  }
  let comments = 0;
  for (let offset = 0; offset < items.length; offset += 50) {
    // eslint-disable-next-line no-await-in-loop -- Source IDs are bounded per request, preserving the same digest for parents and retries.
    const batch = await Promise.all(
      items.slice(offset, offset + 50).map(async (item) => ({
        ...item,
        sourceId: await getBoundedImportSourceKey(item.sourceId),
        ...(item.parentSourceId
          ? {
              parentSourceId: await getBoundedImportSourceKey(
                item.parentSourceId,
              ),
            }
          : {}),
      })),
    );
    try {
      // eslint-disable-next-line no-await-in-loop -- Parent comments must commit before child replies; bounded batches make retries truthful.
      const raw = await post<{ items: CommentItem[] }, unknown>(
        "stories/import/comments",
        { items: batch },
        ctx,
      );
      const response = commentResponseSchema.parse(raw);
      if (response.error?.message || !response.data)
        throw new Error(
          response.error?.message || "The comment batch could not be imported.",
        );
      const results = new Map(
        response.data.items.map((item) => [
          `${item.storyId}:${item.sourceId}`,
          item,
        ]),
      );
      for (const item of batch) {
        const result = results.get(`${item.storyId}:${item.sourceId}`);
        if (!result) {
          unresolvedComments += 1;
          commentIssues.push(
            `${item.sourceId}: The server did not confirm this comment. Retry safely.`,
          );
          continue;
        }
        if (result.error) {
          unresolvedComments += 1;
          commentIssues.push(`${result.sourceId}: ${result.error.message}`);
        } else if (result.commentId) comments += 1;
        else {
          unresolvedComments += 1;
          commentIssues.push(
            `${result.sourceId}: No committed comment was confirmed.`,
          );
        }
      }
    } catch (error) {
      unresolvedComments += batch.length;
      commentIssues.push(
        error instanceof Error ? error.message : "A comment batch failed.",
      );
    }
    onProgress(
      items.length
        ? 95 + Math.round(((offset + batch.length) / items.length) * 3)
        : 98,
    );
  }
  return { comments, unresolvedComments, commentIssues };
};
