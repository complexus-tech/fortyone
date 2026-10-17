import { renderDocumentComment } from "./document-comment-markdown";

export function DocumentCommentContent({ body }: { body: string }) {
  return (
    <div
      className="document-comment-content text-foreground [&_a]:text-primary [&_blockquote]:border-border [&_code]:bg-surface-muted [&_pre]:bg-surface-muted [&_td]:border-border [&_th]:border-border text-base leading-6 [overflow-wrap:anywhere] break-words [&_a]:underline [&_blockquote]:my-2 [&_blockquote]:border-l-2 [&_blockquote]:pl-3 [&_code]:rounded [&_code]:px-1 [&_code]:text-[0.95rem] [&_h1]:my-2 [&_h1]:text-lg [&_h1]:font-semibold [&_h2]:my-2 [&_h2]:text-lg [&_h2]:font-semibold [&_h3]:my-2 [&_h3]:font-semibold [&_li]:my-1 [&_ol]:my-2 [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:my-2 [&_p:first-child]:mt-0 [&_p:last-child]:mb-0 [&_pre]:my-2 [&_pre]:overflow-x-auto [&_pre]:rounded-lg [&_pre]:p-3 [&_table]:block [&_table]:overflow-x-auto [&_td]:border [&_td]:px-2 [&_th]:border [&_th]:px-2 [&_ul]:my-2 [&_ul]:list-disc [&_ul]:pl-5"
      dangerouslySetInnerHTML={{ __html: renderDocumentComment(body) }}
    />
  );
}
