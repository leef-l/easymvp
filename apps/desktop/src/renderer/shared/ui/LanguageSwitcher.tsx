import { useTranslation } from "react-i18next";

const LANG_KEY = "easymvp.desktop.language";

export function LanguageSwitcher() {
  const { i18n } = useTranslation();
  const current = i18n.language;

  function switchTo(lng: string) {
    i18n.changeLanguage(lng);
    localStorage.setItem(LANG_KEY, lng);
  }

  return (
    <div className="lang-switcher">
      <button
        className={`lang-btn${current === "zh-CN" ? " is-active" : ""}`}
        onClick={() => switchTo("zh-CN")}
      >
        中文
      </button>
      <button
        className={`lang-btn${current.startsWith("en") ? " is-active" : ""}`}
        onClick={() => switchTo("en")}
      >
        EN
      </button>
    </div>
  );
}
