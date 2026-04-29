import React from "react";
import ReactDOM from "react-dom/client";
import { AppRouter } from "./app/router";
import "./shared/lib/i18n";
import "./styles/global.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <AppRouter />
  </React.StrictMode>,
);
