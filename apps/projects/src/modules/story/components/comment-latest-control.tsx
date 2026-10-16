import type { RefObject } from "react";
import { useEffect, useState } from "react";
import { Button, Flex } from "ui";

const observeCommentVisibility = (
  container: HTMLDivElement,
  commentId: string,
  onVisibilityChange: (visible: boolean) => void,
) => {
  let target: Element | null = null;
  const observer = new IntersectionObserver(([entry]) => {
    if (entry.target === target) onVisibilityChange(entry.isIntersecting);
  });
  const observeTarget = () => {
    const row = document.getElementById(`comment-${commentId}`);
    const next = row?.querySelector("[data-comment-anchor]") ?? row;
    if (target === next) return;
    observer.disconnect();
    target = next;
    if (target) observer.observe(target);
  };
  // A latest reply can be inside a collapsed thread or arrive on another page.
  // Rebind when that row mounts, without changing the visit's unread boundary.
  const mutations = new MutationObserver(() => {
    observeTarget();
    if (!target) onVisibilityChange(false);
  });
  observeTarget();
  mutations.observe(container, { childList: true, subtree: true });
  return () => {
    observer.disconnect();
    mutations.disconnect();
  };
};

export const CommentLatestControl = ({
  commentId,
  containerRef,
  onJump,
}: {
  commentId: string;
  containerRef: RefObject<HTMLDivElement | null>;
  onJump: () => void;
}) => {
  const [visibleCommentId, setVisibleCommentId] = useState<string | null>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    return observeCommentVisibility(container, commentId, (visible) => {
      setVisibleCommentId(visible ? commentId : null);
    });
  }, [commentId, containerRef]);

  if (visibleCommentId === commentId) return null;

  return (
    <Flex align="center" className="mb-3 ml-9" justify="end">
      <Button
        className="px-2 text-[0.95rem]"
        color="tertiary"
        onClick={onJump}
        size="sm"
        variant="naked"
      >
        Jump to latest
      </Button>
    </Flex>
  );
};
