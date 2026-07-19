import { afterEach, describe, expect, it, vi } from "vitest";
import { apiClient } from "@/lib/api-client";

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
    ...init,
  });
}

describe("apiClient", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("includes credentials on fetch requests", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ status: "ok" }));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.getHealth();

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/health",
      expect.objectContaining({ credentials: "include" }),
    );
  });

  it("posts structured vacation settings", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({ email: "alice@local.test", enabled: true, status: "enabled" }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.setVacation("alice@local.test", {
      subject: "Away",
      body: "Back soon",
      enabled: true,
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/mail/vacation/alice%40local.test",
      expect.objectContaining({
        body: JSON.stringify({ subject: "Away", body: "Back soon", enabled: true }),
        credentials: "include",
        method: "POST",
      }),
    );
  });
});
