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
  PublicPortalViewer,
} from "./types";

const DEFAULT_FILTERS: PublicPortalFilters = {
  search: "",
  sort: "top",
};

export const PublicPortalRequestsPage = ({
  initialFilters = DEFAULT_FILTERS,
  portal,
  viewer,
}: {
  initialFilters?: PublicPortalFilters;
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
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
    <PublicPortalShell activeTab="feedback" portal={portal} viewer={viewer}>
      <Box className="mx-auto grid w-full max-w-[78rem] gap-10 px-4 py-8 md:grid-cols-[minmax(0,1fr)_19rem] md:px-6">
        <Box className="min-h-0">
          <PublicFeedbackList
            filters={filters}
            initialFilters={initialFilters}
            onFiltersChange={updateFilters}
            portal={portal}
          />
        </Box>
        <PublicPortalSidebar
          onBoardSelect={(boardId) => {
            updateFilters({ boardId });
          }}
          portal={portal}
          selectedBoardId={filters.boardId}
          viewer={viewer}
        />
      </Box>
    </PublicPortalShell>
  );
};
