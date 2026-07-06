export type CalendarConnection = {
  id: string;
  provider: string;
  connectedEmail: string;
  timezone: string;
  scopes: string[];
  canReadEventDetails: boolean;
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
