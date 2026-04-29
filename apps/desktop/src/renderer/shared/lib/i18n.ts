import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import zhCN from "../../locales/zh-CN.json";
import en from "../../locales/en.json";

const LANG_KEY = "easymvp.desktop.language";

const resources = {
  "zh-CN": { translation: zhCN },
  en: { translation: en },
};

const savedLang = typeof localStorage !== "undefined"
  ? localStorage.getItem(LANG_KEY) || "zh-CN"
  : "zh-CN";

const finalLang = savedLang.startsWith("zh") ? "zh-CN" : savedLang;

i18n
  .use(initReactI18next)
  .init({
    resources,
    lng: "zh-CN",
    fallbackLng: "en",
    interpolation: {
      escapeValue: false,
    },
  });

export default i18n;
