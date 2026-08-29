"use client";

import { useEffect, useState } from "react";
import { Box } from "ui";
import { PublicPortalShell } from "./portal-shell";
import { PublicPortalSidebar } from "./sidebar";
import { PublicFeedbackList } from "./feedback-list";
import {
  parsePublicPortalFilters,
  toPublicPortalSearchParams,
} from "./query-params";
import type {
  PublicPortal,
  PublicPortalFilters,
  PublicPortalParticipant,
} from "./types";
import { anonymousPublicPortalParticipant } from "./participant";

const DEFAULT_FILTERS: PublicPortalFilters = {
  search: "",
  sort: "top",
  status: "active",
};

export const PublicPortalRequestsPage = ({
  initialFeedbackComposerOpen = false,
  initialFilters = DEFAULT_FILTERS,
  participant = anonymousPublicPortalParticipant,
  portal,
}: {
  initialFeedbackComposerOpen?: boolean;
  initialFilters?: PublicPortalFilters;
  participant?: PublicPortalParticipant;
  portal: PublicPortal;
}) => {
  const [filters, setFilters] = useState(initialFilters);

  const updateFilters = (updates: Partial<PublicPortalFilters>) => {
    const nextFilters = { ...filters, ...updates };
    const params = toPublicPortalSearchParams(nextFilters);
    const query = params.toString();

    window.history.replaceState(
      window.history.state,
      "",
      `${window.location.pathname}${query ? `?${query}` : ""}${window.location.hash}`,
    );
    setFilters(nextFilters);
  };

  useEffect(() => {
    const restoreFilters = () => {
      setFilters(
        parsePublicPortalFilters(new URLSearchParams(window.location.search)),
      );
    };

    window.addEventListener("popstate", restoreFilters);
    return () => {
      window.removeEventListener("popstate", restoreFilters);
    };
  }, []);

  return (
    <PublicPortalShell
      activeTab="feedback"
      participant={participant}
      portal={portal}
    >
      <Box className="mx-auto grid w-full max-w-[78rem] gap-10 px-4 py-8 md:grid-cols-[minmax(0,1fr)_19rem] md:px-6">
        <Box className="min-h-0">
          <PublicFeedbackList
            filters={filters}
            initialFilters={initialFilters}
            onFiltersChange={updateFilters}
            participant={participant}
            portal={portal}
          />
        </Box>
        <PublicPortalSidebar
          initialFeedbackComposerOpen={initialFeedbackComposerOpen}
          onBoardSelect={(boardId) => {
            updateFilters({ boardId });
          }}
          participant={participant}
          portal={portal}
          selectedBoardId={filters.boardId}
        />
      </Box>
    </PublicPortalShell>
  );
};
