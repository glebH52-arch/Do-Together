export type Locale = "ru" | "en";

const dictionaries = {
  ru: {
    projects: "Проекты", invites: "Приглашения", notifications: "Уведомления", settings: "Настройки",
    history: "История", addTask: "Добавить задачу", team: "Команда", logout: "Выйти", close: "Закрыть",
    emptyInvites: "Новых приглашений нет", emptyNotifications: "Новых уведомлений нет",
    accept: "Принять", decline: "Отклонить", language: "Язык", theme: "Тема", profile: "Профиль",
    light: "Светлая", dark: "Тёмная", system: "Системная", save: "Сохранить", removeAvatar: "Удалить фото",
  },
  en: {
    projects: "Projects", invites: "Invites", notifications: "Notifications", settings: "Settings",
    history: "History", addTask: "Add task", team: "Team", logout: "Sign out", close: "Close",
    emptyInvites: "No pending invites", emptyNotifications: "No new notifications",
    accept: "Accept", decline: "Decline", language: "Language", theme: "Theme", profile: "Profile",
    light: "Light", dark: "Dark", system: "System", save: "Save", removeAvatar: "Remove photo",
  },
} as const;

export function useCopy(locale: Locale) { return dictionaries[locale]; }
export const roleLabel = (v: string, locale: Locale) => ({ru:{creator:"Владелец",admin:"Модератор",member:"Участник"},en:{creator:"Owner",admin:"Moderator",member:"Member"}}[locale] as Record<string,string>)[v] ?? v;
export const statusLabel = (v: string, locale: Locale) => ({ru:{todo:"К выполнению",in_progress:"В работе",review:"На проверке",done:"Готово"},en:{todo:"To do",in_progress:"In progress",review:"Review",done:"Done"}}[locale] as Record<string,string>)[v] ?? v;
export const fmt = (value: string | null | undefined, locale: Locale, dateOnly=false) => value ? new Intl.DateTimeFormat(locale === "ru" ? "ru-RU" : "en-US", dateOnly ? {dateStyle:"medium"}:{dateStyle:"medium",timeStyle:"short"}).format(new Date(value)) : "—";
