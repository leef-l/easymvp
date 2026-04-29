import { render, screen, waitFor } from "@testing-library/react";
import { AuditPage } from "@/modules/audit/pages/AuditPage";
import { BrowserRouter } from "react-router-dom";

Object.defineProperty(window, "desktopBridge", {
  value: { platform: "test", version: "0.0.0-test", coreBaseUrl: "" },
  writable: true,
});

const originalFetch = globalThis.fetch;

const mockAuditData = {
  data: {
    items: [
      {
        id: "audit-1",
        event_type: "plan_compiled",
        actor_kind: "system",
        summary: "Plan compiled successfully",
        created_at: "2026-04-25T10:00:00Z",
        payload_json: '{"task_id":"task-1"}',
      },
    ],
    refresh_hint: "latest",
  },
};

describe("AuditPage", () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = typeof input === "string" ? input : input.toString();
      if (url.includes("/audit-logs")) {
        return new Response(JSON.stringify(mockAuditData), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }
      return new Response("{}", { status: 404 });
    }) as typeof fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it("renders audit page with data", async () => {
    render(
      <BrowserRouter>
        <AuditPage />
      </BrowserRouter>,
    );
    await waitFor(() => {
      expect(screen.getByText("Plan compiled successfully")).toBeTruthy();
    });
  });

  it("does not use window.location.reload for retry", () => {
    const reloadSpy = vi.fn();
    Object.defineProperty(window, "location", {
      value: { ...window.location, reload: reloadSpy },
      writable: true,
    });
    render(
      <BrowserRouter>
        <AuditPage />
      </BrowserRouter>,
    );
    expect(reloadSpy).not.toHaveBeenCalled();
  });
});
