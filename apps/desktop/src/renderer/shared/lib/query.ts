import { useEffect, useState } from "react";

export type QueryState<T> = {
  data: T | null;
  error: string;
  loading: boolean;
  refreshing: boolean;
  stale: boolean;
};

export function useQuery<T>(load: () => Promise<T>, deps: readonly unknown[]): QueryState<T> {
  const [state, setState] = useState<QueryState<T>>({
    data: null,
    error: "",
    loading: true,
    refreshing: false,
    stale: false,
  });

  useEffect(() => {
    let active = true;
    setState((current) => ({
      data: current.data,
      error: "",
      loading: current.data === null,
      refreshing: current.data !== null,
      stale: current.data !== null,
    }));

    void load()
      .then((data) => {
        if (!active) {
          return;
        }
        setState({
          data,
          error: "",
          loading: false,
          refreshing: false,
          stale: false,
        });
      })
      .catch((error: unknown) => {
        if (!active) {
          return;
        }
        setState((current) => ({
          data: current.data,
          error: error instanceof Error ? error.message : "Request failed",
          loading: false,
          refreshing: false,
          stale: current.data !== null,
        }));
      });

    return () => {
      active = false;
    };
  }, deps);

  return state;
}
