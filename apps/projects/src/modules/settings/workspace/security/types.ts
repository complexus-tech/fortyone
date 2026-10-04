import { z } from "zod";

const date = z.string().datetime({ offset: true });
export const policySchema = z.object({
  allowedDomains: z.array(z.string()),
  allowGuests: z.boolean(),
  maxSessionAgeHours: z.number().int().min(0).max(720),
  version: z.number().int().nonnegative(),
  updatedAt: date.nullable(),
});
export type SecurityPolicy = z.infer<typeof policySchema>;
export type PolicyUpdate = Omit<SecurityPolicy, "version" | "updatedAt"> & {
  expectedVersion: number;
};
export const browserSessionSchema = z.object({
  id: z.string().uuid(),
  userId: z.string().uuid(),
  name: z.string(),
  username: z.string().optional(),
  browserName: z.string().nullable().optional(),
  email: z.string(),
  role: z.string(),
  authenticatedAt: date,
  lastSeenAt: date,
  expiresAt: date,
  revokedAt: date.nullable(),
  revokedBefore: date.nullable(),
  current: z.boolean(),
});
export type BrowserSession = z.infer<typeof browserSessionSchema>;
export const sessionListSchema = z.object({
  items: browserSessionSchema.array(),
  hasMore: z.boolean(),
});
export const auditEventSchema = z.object({
  id: z.string().uuid(),
  source: z.string(),
  actorId: z.string().uuid().nullable(),
  actorType: z.string(),
  resourceType: z.string(),
  resourceId: z.string().uuid().nullable(),
  operation: z.string(),
  metadata: z.record(z.string(), z.unknown()),
  createdAt: date,
});
export type AuditEvent = z.infer<typeof auditEventSchema>;
export const auditPageSchema = z.object({
  items: auditEventSchema.array(),
  nextCursor: z.string(),
});
export type AuditFilters = {
  actorId?: string;
  resourceType?: string;
  resourceId?: string;
  from?: string;
  to?: string;
};
export const sessionStatus = (session: BrowserSession) => {
  if (
    session.revokedAt ||
    (session.revokedBefore &&
      Date.parse(session.authenticatedAt) <= Date.parse(session.revokedBefore))
  )
    return "Revoked";
  return Date.parse(session.expiresAt) <= Date.now() ? "Expired" : "Active";
};
