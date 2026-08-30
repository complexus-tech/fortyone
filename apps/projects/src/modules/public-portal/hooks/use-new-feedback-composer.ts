"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Underline from "@tiptap/extension-underline";
import { useEditor } from "@tiptap/react";
import { toast } from "sonner";
import { useDebouncedCallback } from "@/hooks/debounce";
import { createRichTextStarterKit } from "@/lib/tiptap/starter-kit";
import { getCurrentFeedbackGuestAction } from "../actions";
import { getAnonymousFeedbackTrackingUrl } from "../anonymous-tracking";
import { useSimilarPublicFeedback } from "../client-query";
import {
  useCreateAnonymousPublicFeedback,
  useCreatePublicFeedback,
} from "../feedback-mutations";
import { isContactableParticipant } from "../participant";
import { NEW_FEEDBACK_QUERY_PARAM } from "../query-params";
import type { PublicPortal, PublicPortalParticipant } from "../types";
import { getRequestPathBySlug } from "../utils";

type ComposerStep = "draft" | "participation" | "verification";

export const useNewFeedbackComposer = ({
  initialOpen = false,
  participant,
  portal,
}: {
  initialOpen?: boolean;
  participant: PublicPortalParticipant;
  portal: PublicPortal;
}) => {
  const router = useRouter();
  const [open, setOpen] = useState(initialOpen);
  const [composerStep, setComposerStep] = useState<ComposerStep>("draft");
  const [lockedParticipant, setLockedParticipant] =
    useState<PublicPortalParticipant>(participant);
  const [isCheckingGuestSession, setIsCheckingGuestSession] = useState(false);
  const [title, setTitle] = useState("");
  const [anonymousSubmission, setAnonymousSubmission] = useState<{
    trackingUrl: string;
  } | null>(null);
  const titleRef = useRef("");
  const [similarityInput, setSimilarityInput] = useState({
    description: "",
    title: "",
  });
  const [boardId, setBoardId] = useState(
    portal.boards.length === 1 ? portal.boards[0]?.id ?? "" : "",
  );
  const createFeedback = useCreatePublicFeedback({
    participant: lockedParticipant,
    portal,
  });
  const createAnonymousFeedback = useCreateAnonymousPublicFeedback({ portal });
  const isSubmitting =
    createFeedback.isPending ||
    createAnonymousFeedback.isPending ||
    isCheckingGuestSession;
  const selectedBoard = portal.boards.find((board) => board.id === boardId);
  const { callback: checkForSimilarFeedback, cancel: cancelSimilarityCheck } =
    useDebouncedCallback(setSimilarityInput, 300);
  const descriptionEditor = useEditor({
    content: "",
    editable: true,
    editorProps: {
      attributes: {
        "aria-label": "Feedback description",
        class: "min-h-24 outline-none",
      },
    },
    extensions: [
      createRichTextStarterKit(),
      Underline,
      Link.configure({ autolink: true }),
      Placeholder.configure({
        placeholder: "Describe the feedback, context, or expected outcome...",
      }),
    ],
    immediatelyRender: false,
    onUpdate: ({ editor }) => {
      checkForSimilarFeedback({
        description: editor.getText(),
        title: titleRef.current,
      });
    },
  });
  const similarFeedback = useSimilarPublicFeedback({
    description: similarityInput.description,
    portalSlug: portal.slug,
    title: open ? similarityInput.title : "",
  });
  const similarFeedbackItems =
    title.trim() === similarityInput.title.trim()
      ? similarFeedback.data ?? []
      : [];
  const blockingMatch = similarFeedbackItems.find((item) => item.isDuplicate);

  const clearInitialOpenIntent = () => {
    if (!initialOpen) return;

    const url = new URL(window.location.href);
    url.searchParams.delete(NEW_FEEDBACK_QUERY_PARAM);
    window.history.replaceState(window.history.state, "", url);
  };

  const close = () => {
    clearInitialOpenIntent();
    cancelSimilarityCheck();
    setAnonymousSubmission(null);
    setComposerStep("draft");
    setLockedParticipant(participant);
    setOpen(false);
  };

  const resetDraft = () => {
    setTitle("");
    titleRef.current = "";
    setSimilarityInput({ description: "", title: "" });
    descriptionEditor?.commands.setContent("");
  };

  const openExistingFeedback = (slug: string, isDuplicate = false) => {
    close();
    router.push(getRequestPathBySlug(portal, slug));
    if (isDuplicate) {
      toast.info("This feedback was already reported", {
        description: "Add your context as a comment on the existing feedback.",
      });
    }
  };

  const getDraftInput = () => ({
    boardId,
    description: descriptionEditor?.getText() ?? "",
    portalSlug: portal.slug,
    title,
  });

  const submitAnonymously = () => {
    const input = getDraftInput();
    createAnonymousFeedback.mutate(input, {
      onSuccess: (result) => {
        const trackingUrl = getAnonymousFeedbackTrackingUrl(
          portal,
          result.request,
        );
        resetDraft();
        setAnonymousSubmission({ trackingUrl });
        toast.success("Feedback submitted anonymously");
      },
    });
  };

  const submitAsContactableParticipant = (
    activeParticipant: Exclude<PublicPortalParticipant, { kind: "anonymous" }>,
  ) => {
    const input = getDraftInput();
    createFeedback.mutate(
      { ...input, participant: activeParticipant },
      {
        onError: async () => {
          setTitle(input.title);
          titleRef.current = input.title;
          descriptionEditor?.commands.setContent(input.description);
          setComposerStep("draft");
          setOpen(true);
          const refreshed = await similarFeedback.refetch();
          const duplicate = refreshed.data?.find((item) => item.isDuplicate);
          if (duplicate) openExistingFeedback(duplicate.slug, true);
        },
        onSuccess: () => {
          close();
          resetDraft();
          router.refresh();
          toast.success("Feedback submitted", {
            description:
              activeParticipant.kind === "account"
                ? undefined
                : "You are following this feedback and can receive meaningful updates.",
          });
        },
      },
    );
  };

  const continueWithEmail = async () => {
    if (isContactableParticipant(lockedParticipant)) {
      submitAsContactableParticipant(lockedParticipant);
      return;
    }

    setIsCheckingGuestSession(true);
    const response = await getCurrentFeedbackGuestAction(portal.slug);
    setIsCheckingGuestSession(false);
    if (response.data?.participant) {
      setLockedParticipant(response.data.participant);
      submitAsContactableParticipant(response.data.participant);
      return;
    }
    setComposerStep("verification");
  };

  const submit = () => {
    if (blockingMatch) {
      openExistingFeedback(blockingMatch.slug, true);
      return;
    }
    if (portal.participationMode === "anonymous_allowed") {
      setComposerStep("participation");
      return;
    }
    if (!isContactableParticipant(lockedParticipant)) {
      void continueWithEmail();
      return;
    }
    submitAsContactableParticipant(lockedParticipant);
  };

  const handleDialogOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && isSubmitting) return;

    setOpen(nextOpen);
    if (!nextOpen) {
      clearInitialOpenIntent();
      cancelSimilarityCheck();
      setAnonymousSubmission(null);
      setComposerStep("draft");
      setLockedParticipant(participant);
    }
  };

  const openComposer = () => {
    setAnonymousSubmission(null);
    setComposerStep("draft");
    setLockedParticipant(participant);
    setOpen(true);
    checkForSimilarFeedback({
      description: descriptionEditor?.getText() ?? "",
      title: titleRef.current,
    });
  };

  const updateTitle = (nextTitle: string) => {
    setTitle(nextTitle);
    titleRef.current = nextTitle;
    checkForSimilarFeedback({
      description: descriptionEditor?.getText() ?? "",
      title: nextTitle,
    });
  };

  return {
    anonymousSubmission,
    blockingMatch,
    boardId,
    close,
    composerStep,
    continueWithEmail,
    descriptionEditor,
    handleDialogOpenChange,
    isSubmitting,
    lockedParticipant,
    open,
    openComposer,
    openExistingFeedback,
    selectedBoard,
    setBoardId,
    setComposerStep,
    setLockedParticipant,
    similarFeedbackItems,
    submit,
    submitAnonymously,
    submitAsContactableParticipant,
    title,
    updateTitle,
  };
};

export type NewFeedbackComposerState = ReturnType<
  typeof useNewFeedbackComposer
>;
