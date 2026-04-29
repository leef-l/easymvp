import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

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
  const { t } = useTranslation();
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
          <span className="status-pill">{t("query.loading")}</span>
        </div>
        <p className="muted-copy">{t("query.loadingDesc")}</p>
      </section>
    );
  }

  if (error) {
    return (
      <section className="data-panel is-error">
        <div className="panel-header">
          <h3>{title}</h3>
          <span className="status-pill">{t("query.error")}</span>
        </div>
        <p className="error-copy">{error}</p>
        {recoveryMessage ? <p className="muted-copy">{recoveryMessage}</p> : null}
        {onRetry || (secondaryActionLabel && onSecondaryAction) ? (
          <div className="action-row">
            {onRetry ? (
              <button className="secondary-button" onClick={onRetry}>
                {t("query.retry")}
              </button>
            ) : null}
            {secondaryActionLabel && onSecondaryAction ? (
              <button className="secondary-button" onClick={onSecondaryAction}>
                {secondaryActionLabel}
              </button>
            ) : null}
          </div>
        ) : null}
        {stale ? <p className="muted-copy">{t("query.staleDesc")}</p> : null}
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
            <span className="status-pill">{t("query.refreshing")}</span>
          </div>
          <p className="muted-copy">{t("query.refreshingDesc")}</p>
        </section>
      ) : null}
      {children}
    </>
  );
}
