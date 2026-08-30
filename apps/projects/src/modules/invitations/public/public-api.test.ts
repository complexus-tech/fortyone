/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync, readdirSync, statSync } from "node:fs";
import path from "node:path";
import * as invitationClient from "./client";
import * as onboardingCapabilities from "./onboarding";
import type { Invitation, NewInvitation } from "./types";

const modulesRoot = path.resolve(__dirname, "../..");

const walkProductionSources = (directory: string): string[] =>
  readdirSync(directory).flatMap((entry) => {
    const entryPath = path.join(directory, entry);
    if (statSync(entryPath).isDirectory()) {
      return walkProductionSources(entryPath);
    }
    return /\.(?:ts|tsx)$/.test(entryPath) && !/\.test\.(?:ts|tsx)$/.test(entry)
      ? [entryPath]
      : [];
  });

describe("invitations public API", () => {
  it("exposes only the intended browser hooks and onboarding capabilities", () => {
    expect(Object.keys(invitationClient).sort()).toEqual([
      "useAcceptInvitationMutation",
      "useMyInvitations",
      "usePendingInvitations",
      "useRevokeInvitationMutation",
    ]);
    expect(Object.keys(onboardingCapabilities).sort()).toEqual([
      "acceptInvitation",
      "inviteOnboardingMembers",
    ]);
  });

  it("keeps its public domain types transport-only", () => {
    const invitation = {
      createdAt: "2026-08-29T08:00:00.000Z",
      email: "ada@example.com",
      expiresAt: "2026-08-30T08:00:00.000Z",
      id: "invitation-1",
      inviterId: "inviter-1",
      role: "admin",
      teamIds: ["team-1"],
      updatedAt: "2026-08-29T08:00:00.000Z",
      workspaceColor: "#111111",
      workspaceId: "workspace-1",
      workspaceName: "Complexus",
      workspaceSlug: "complexus",
    } satisfies Invitation;
    const newInvitation = {
      email: invitation.email,
      role: invitation.role,
      teamIds: invitation.teamIds,
    } satisfies NewInvitation;

    expect(newInvitation).toEqual({
      email: "ada@example.com",
      role: "admin",
      teamIds: ["team-1"],
    });
  });

  it("marks the server entrypoint server-only and keeps client commands free of request headers", () => {
    const serverSource = readFileSync(
      path.join(__dirname, "server.ts"),
      "utf8",
    );
    const acceptSource = readFileSync(
      path.join(__dirname, "../actions/accept-invitation.ts"),
      "utf8",
    );

    expect(serverSource).toContain('import "server-only";');
    expect(acceptSource).not.toContain("next/headers");
    expect(acceptSource).not.toContain("getCookieHeader");
  });

  it("routes every cross-module invitation dependency through public entrypoints", () => {
    const privateImports: string[] = [];

    for (const moduleDirectory of readdirSync(modulesRoot)) {
      if (moduleDirectory === "invitations") continue;
      const sourceDirectory = path.join(modulesRoot, moduleDirectory);
      if (!statSync(sourceDirectory).isDirectory()) continue;

      for (const sourcePath of walkProductionSources(sourceDirectory)) {
        const source = readFileSync(sourcePath, "utf8");
        if (/from ["']@\/modules\/invitations\/(?!public\/)/.test(source)) {
          privateImports.push(path.relative(modulesRoot, sourcePath));
        }
      }
    }

    expect(privateImports).toEqual([]);
  });
});
