import type { ReactNode } from "react";

type QueryPanelProps = {
  loading: boolean;
  refreshing?: boolean;
  stale?: boolean;
  error: string;
  title: string;
  onRetry?: () => void;
  secondaryActionLabel?: string;
  onSecondaryAction?: () => void;
  recoveryMessage?: string;
  children: ReactNode;
};

export function QueryPanel(props: QueryPanelProps) {
  const {
    loading,
    refreshing,
    stale,
    error,
    title,
    onRetry,
    secondaryActionLabel,
    onSecondaryAction,
    recoveryMessage,
    children,
  } = props;

  if (loading) {
    return (
      <section className="data-panel">
        <div className="panel-header">
          <h3>{title}</h3>
          <span className="status-pill">Loading</span>
        </div>
        <p className="muted-copy">Fetching latest data from local core service.</p>
      </section>
    );
  }

  if (error) {
    return (
      <section className="data-panel is-error">
        <div className="panel-header">
          <h3>{title}</h3>
          <span className="status-pill">Error</span>
        </div>
        <p className="error-copy">{error}</p>
        {recoveryMessage ? <p className="muted-copy">{recoveryMessage}</p> : null}
        {onRetry || (secondaryActionLabel && onSecondaryAction) ? (
          <div className="action-row">
            {onRetry ? (
              <button className="secondary-button" onClick={onRetry}>
                Retry
              </button>
            ) : null}
            {secondaryActionLabel && onSecondaryAction ? (
              <button className="secondary-button" onClick={onSecondaryAction}>
                {secondaryActionLabel}
              </button>
            ) : null}
          </div>
        ) : null}
        {stale ? <p className="muted-copy">Showing the last successful result while the latest refresh failed.</p> : null}
        {children}
      </section>
    );
  }

  return (
    <>
      {refreshing ? (
        <section className="data-panel">
          <div className="panel-header">
            <h3>{title}</h3>
            <span className="status-pill">Refreshing</span>
          </div>
          <p className="muted-copy">Keeping the last successful result while fetching a newer snapshot.</p>
        </section>
      ) : null}
      {children}
    </>
  );
}
