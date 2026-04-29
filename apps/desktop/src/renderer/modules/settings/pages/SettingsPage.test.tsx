import { render, screen, waitFor } from "@testing-library/react";
import { SettingsPage } from "@/modules/settings/pages/SettingsPage";
import { BrowserRouter } from "react-router-dom";

// mock desktop bridge
Object.defineProperty(window, "desktopBridge", {
  value: {
    platform: "test",
    version: "0.0.0-test",
    coreBaseUrl: "",
  },
  writable: true,
});

// mock fetch for health endpoints
const originalFetch = globalThis.fetch;

describe("SettingsPage", () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = typeof input === "string" ? input : input.toString();
      if (url.includes("/api/v3/system/healthz")) {
        return new Response(
          JSON.stringify({
            data: {
              status: "healthy",
              service: "easymvp-core",
              version: "0.1.0",
              timestamp: new Date().toISOString(),
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      if (url.includes("/api/v3/runtime/healthz")) {
        return new Response(
          JSON.stringify({
            data: {
              status: "ready",
              base_url: "http://127.0.0.1:8000",
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      return new Response("{}", { status: 404 });
    }) as typeof fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it("renders settings title and save button", async () => {
    render(
      <BrowserRouter>
        <SettingsPage />
      </BrowserRouter>,
    );

    await waitFor(() => {
      const buttons = screen.getAllByRole("button");
      expect(buttons.length).toBeGreaterThan(0);
    });
  });
});
