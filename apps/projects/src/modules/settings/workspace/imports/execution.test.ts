/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { webcrypto } from "node:crypto";
import { zodTextFormat } from "openai/helpers/zod";
import type { Member } from "@/types";
import type { State } from "@/types/states";
import { importAnalysisSchema } from "./schema";
import {
  getBoundedImportSourceKey,
  resolveImportAssignee,
  resolveImportStatus,
} from "./execution";

const status = (
  id: string,
  name: string,
  category: State["category"],
  isDefault = false,
): State => ({
  id,
  name,
  category,
  isDefault,
  orderIndex: 1,
  color: "#000000",
  teamId: "team",
  workspaceId: "workspace",
  createdAt: "2026-01-01",
  updatedAt: "2026-01-01",
});

describe("work import execution mapping", () => {
  it("keeps the AI analysis contract convertible to a strict response format", () => {
    Object.defineProperty(globalThis, "structuredClone", {
      configurable: true,
      value: <T>(value: T) => JSON.parse(JSON.stringify(value)) as T,
    });
    expect(() =>
      zodTextFormat(importAnalysisSchema, "work_import_analysis"),
    ).not.toThrow();
  });

  it("prefers an exact status before mapping Jira status categories", () => {
    const statuses = [
      status("default-started", "Started", "started", true),
      status("exact", "In Progress", "started"),
    ];

    expect(resolveImportStatus("Doing now", statuses)?.id).toBe(
      "default-started",
    );
    expect(resolveImportStatus("In Progress", statuses)?.id).toBe("exact");
  });

  it("maps completed Jira statuses and falls back to the default unstarted state", () => {
    const statuses = [
      status("todo", "Todo", "unstarted", true),
      status("done", "Finished", "completed", true),
    ];

    expect(resolveImportStatus("Resolved", statuses)?.id).toBe("done");
    expect(resolveImportStatus("Unknown state", statuses)?.id).toBe("todo");
  });

  it("only assigns an active team member with an exact email match", () => {
    const member = {
      id: "member-1",
      email: "Owner@Example.com",
      isActive: true,
    } as Member;

    expect(resolveImportAssignee("owner@example.com", [member])?.id).toBe(
      "member-1",
    );
    expect(
      resolveImportAssignee("different@example.com", [member]),
    ).toBeUndefined();
  });

  it("keeps ordinary Jira keys readable and hashes unsafe or oversized source keys", async () => {
    Object.defineProperty(globalThis.crypto, "subtle", {
      configurable: true,
      value: webcrypto.subtle,
    });

    await expect(getBoundedImportSourceKey(" JIRA-42 ")).resolves.toBe(
      "JIRA-42",
    );

    const unsafe = await getBoundedImportSourceKey("row\n42");
    const oversized = await getBoundedImportSourceKey("🚀".repeat(100));

    expect(unsafe).toMatch(/^source:[a-f0-9]{64}$/);
    expect(oversized).toMatch(/^source:[a-f0-9]{64}$/);
    await expect(getBoundedImportSourceKey("row\n42")).resolves.toBe(unsafe);
  });
});
