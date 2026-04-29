import { render, screen, waitFor } from "@testing-library/react";
import { DiagnosticsPage } from "@/modules/diagnostics/pages/DiagnosticsPage";
import { BrowserRouter } from "react-router-dom";

Object.defineProperty(window, "desktopBridge", {
  value: {
    platform: "test",
    version: "0.0.0-test",
    coreBaseUrl: "",
    getRuntimeInfo: () =>
      Promise.resolve({
        platform: "test",
        version: "0.0.0-test",
        packaged: false,
        launchMode: "normal",
        coreBaseUrl: "http://127.0.0.1:8000",
        dataDirectory: "/tmp/test",
        core: { reachable: true, status: "ok", httpStatus: 200 },
        coreManager: { enabled: false, mode: "external", status: "disabled" },
        startup: { phase: "ready", pending: false },
      }),
  },
  writable: true,
});

const originalFetch = globalThis.fetch;

describe("DiagnosticsPage", () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = typeof input === "string" ? input : input.toString();
      if (url.includes("/system/healthz")) {
        return new Response(
          JSON.stringify({
            data: { status: "ok", service: "easymvp-core", version: "0.1.0", timestamp: new Date().toISOString() },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      if (url.includes("/runtime/healthz")) {
        return new Response(
          JSON.stringify({ data: { status: "ready", base_url: "http://127.0.0.1:8000" } }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      if (url.includes("/diagnostic-records")) {
        return new Response(
          JSON.stringify({ data: { items: [], linked_runs: [], category_counts: {} } }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      return new Response("{}", { status: 404 });
    }) as typeof fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it("renders diagnostics page with health data", async () => {
    render(
      <BrowserRouter>
        <DiagnosticsPage />
      </BrowserRouter>,
    );
    await waitFor(() => {
      const buttons = screen.getAllByRole("button");
      expect(buttons.length).toBeGreaterThan(0);
    });
  });
});
