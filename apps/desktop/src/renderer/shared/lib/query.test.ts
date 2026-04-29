import { describe, it, expect, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useQuery } from "@/shared/lib/query";

describe("useQuery", () => {
  it("should start in loading state", () => {
    const load = vi.fn().mockResolvedValue("data");
    const { result } = renderHook(() => useQuery(load, []));

    expect(result.current.loading).toBe(true);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBe("");
  });

  it("should resolve to data on success", async () => {
    const load = vi.fn().mockResolvedValue({ id: 1, name: "test" });
    const { result } = renderHook(() => useQuery(load, []));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.data).toEqual({ id: 1, name: "test" });
    expect(result.current.error).toBe("");
    expect(result.current.loading).toBe(false);
  });

  it("should set error on failure", async () => {
    const load = vi.fn().mockRejectedValue(new Error("network failure"));
    const { result } = renderHook(() => useQuery(load, []));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBe("network failure");
    expect(result.current.loading).toBe(false);
  });

  it("should call load once per deps change", async () => {
    const load = vi.fn().mockResolvedValue("data");
    const { rerender } = renderHook(({ deps }) => useQuery(load, deps), {
      initialProps: { deps: [1] as readonly number[] },
    });

    await waitFor(() => expect(load).toHaveBeenCalledTimes(1));

    rerender({ deps: [2] });
    await waitFor(() => expect(load).toHaveBeenCalledTimes(2));

    rerender({ deps: [2] });
    // same deps should not trigger another load
    expect(load).toHaveBeenCalledTimes(2);
  });

  it("should keep stale data while refreshing", async () => {
    let resolve: (value: string) => void = () => {};
    const load = vi.fn().mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r;
        }),
    );

    const { result, rerender } = renderHook(
      ({ deps }) => useQuery(load, deps),
      { initialProps: { deps: [1] as readonly number[] } },
    );

    resolve("first");
    await waitFor(() => expect(result.current.data).toBe("first"));

    rerender({ deps: [2] });
    expect(result.current.stale).toBe(true);
    expect(result.current.data).toBe("first");
    expect(result.current.refreshing).toBe(true);

    resolve("second");
    await waitFor(() => expect(result.current.data).toBe("second"));
    expect(result.current.stale).toBe(false);
    expect(result.current.refreshing).toBe(false);
  });
});
