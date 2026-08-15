export type CalendarConnection = {
  id: string;
  provider: string;
  connectedEmail: string;
  timezone: string;
  scopes: string[];
  canReadEventDetails: boolean;
  canWriteEvents: boolean;
  requiresReauthorization: boolean;
  reauthorizationReason?: "google_calendar_write_scope_required";
  syncStatus: string;
  syncError?: string | null;
  lastSyncedAt?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type CalendarIntegration = {
  connections: CalendarConnection[];
};

export type CreateCalendarConnectSessionResponse = {
  authUrl: string;
};
