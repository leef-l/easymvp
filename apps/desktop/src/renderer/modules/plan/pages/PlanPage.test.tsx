import { render, screen, waitFor } from "@testing-library/react";
import { PlanPage } from "@/modules/plan/pages/PlanPage";
import { BrowserRouter } from "react-router-dom";

Object.defineProperty(window, "desktopBridge", {
  value: { platform: "test", version: "0.0.0-test", coreBaseUrl: "" },
  writable: true,
});

const originalFetch = globalThis.fetch;

const mockPlanData = {
  data: {
    draft: { status: "ready", goal_summary: "Test goal" },
    review: { decision: "approved" },
    compiled: { status: "compiled", compiled_version: 1, risk_summary: "" },
    repair_draft: { id: "", status: "idle" },
    diff_summary: {
      summary: "No changes",
      split_count: 0,
      override_count: 0,
      drop_count: 0,
      review_issue_count: 0,
      items: [],
    },
    task_projection: [],
  },
};

describe("PlanPage", () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = typeof input === "string" ? input : input.toString();
      if (url.includes("/plan-view")) {
        return new Response(JSON.stringify(mockPlanData), {
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

  it("renders plan page with goal summary", async () => {
    render(
      <BrowserRouter>
        <PlanPage />
      </BrowserRouter>,
    );
    await waitFor(() => {
      expect(screen.getByText("Test goal")).toBeTruthy();
    });
  });
});
