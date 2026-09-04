const privateResponseHeaders = { "Cache-Control": "private, no-store" };

export const textResponse = (body: string, status: number) =>
  new Response(body, { headers: privateResponseHeaders, status });

export const jsonResponse = (body: unknown) =>
  Response.json(body, { headers: privateResponseHeaders });
