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

export const chromeCopy = {
  "ru": {
    "myProjects": "Мои проекты",
    "projectsIntro": "Рабочие пространства, к которым у вас есть доступ",
    "createProject": "Создать проект",
    "loadingProjects": "Загружаем проекты…",
    "emptyBoard": "Пустая доска",
    "noProjects": "Пока нет проектов",
    "firstProject": "Создайте первый проект для совместной работы.",
    "active": "Активные",
    "activeStatus": "Активен",
    "archive": "Архив",
    "archived": "Архив",
    "archiveEmpty": "Архив пуст",
    "restore": "Восстановить",
    "edit": "Изменить",
    "archiveAction": "В архив",
    "archiveConfirm": "Архивировать «{name}»?",
    "newProject": "Новый проект",
    "editProject": "Изменить проект",
    "title": "Название",
    "goal": "Цель",
    "cancel": "Отмена",
    "saving": "Сохраняем…",
    "save": "Сохранить",
    "loading": "Загружаем…",
    "members": "Участники",
    "membersIntro": "Имена, роли и доступные действия",
    "invite": "Пригласить",
    "you": "Вы",
    "role": "Роль",
    "remove": "Удалить",
    "inviteMember": "Пригласить участника",
    "inviteNote": "Пользователь должен зарегистрироваться и подтвердить email.",
    "send": "Отправить",
    "sending": "Отправляем…",
    "removeMember": "Удалить {name} из проекта?",
    "inviteMissing": "Пользователь с таким email не найден. Он должен сначала зарегистрироваться.",
    "inviteExists": "Приглашение уже отправлено или пользователь уже состоит в команде.",
    "inviteForbidden": "У вас нет права приглашать в этот проект.",
    "inviteFailed": "Не удалось отправить приглашение. Повторите попытку.",
    "allProjects": "Все проекты",
    "invitesIntro": "Входящие приглашения и история решений",
    "noInvites": "Приглашений нет",
    "sender": "Отправитель",
    "status": "Статус",
    "wait": "Подождите…",
    "pending": "Ожидает ответа",
    "accepted": "Принято",
    "declined": "Отклонено",
    "expired": "Истекло",
    "restoring": "Восстанавливаем сессию…",
    "codeSent": "Код отправлен на указанный email.",
    "sixDigits": "Введите ровно шесть цифр.",
    "welcome": "С возвращением",
    "createAccount": "Создать аккаунт",
    "enterCode": "Введите код из письма",
    "username": "Имя пользователя",
    "password": "Пароль",
    "codeLabel": "Шестизначный код",
    "signIn": "Войти",
    "continue": "Продолжить",
    "verify": "Подтвердить",
    "noAccount": "Нет аккаунта? Создать",
    "haveAccount": "Уже есть аккаунт? Войти",
    "avatarFormat": "Нужен JPEG или PNG до 5 МиБ.",
    "choosePhoto": "Выбрать фото",
    "displayName": "Отображаемое имя",
    "uploadFailed": "Не удалось загрузить фото.",
    "requestFailed": "Не удалось выполнить запрос. Повторите попытку.",
    "assignmentChanged": "Изменён исполнитель задачи",
    "statusChanged": "Изменён статус задачи",
    "dueChanged": "Изменён срок задачи",
    "taskUpdated": "Задача обновлена",
    "invitation": "Приглашение в проект"
  },
  "en": {
    "myProjects": "My projects",
    "projectsIntro": "Workspaces you have access to",
    "createProject": "Create project",
    "loadingProjects": "Loading projects…",
    "emptyBoard": "Empty board",
    "noProjects": "No projects yet",
    "firstProject": "Create your first project to work together.",
    "active": "Active",
    "activeStatus": "Active",
    "archive": "Archive",
    "archived": "Archived",
    "archiveEmpty": "Archive is empty",
    "restore": "Restore",
    "edit": "Edit",
    "archiveAction": "Archive",
    "archiveConfirm": "Archive “{name}”?",
    "newProject": "New project",
    "editProject": "Edit project",
    "title": "Title",
    "goal": "Goal",
    "cancel": "Cancel",
    "saving": "Saving…",
    "save": "Save",
    "loading": "Loading…",
    "members": "Members",
    "membersIntro": "Names, roles and available actions",
    "invite": "Invite",
    "you": "You",
    "role": "Role",
    "remove": "Delete",
    "inviteMember": "Invite a member",
    "inviteNote": "The user must register and verify their email.",
    "send": "Send",
    "sending": "Sending…",
    "removeMember": "Remove {name} from the project?",
    "inviteMissing": "No user with this email was found. They must register first.",
    "inviteExists": "An invitation was already sent or the user is already a member.",
    "inviteForbidden": "You cannot invite members to this project.",
    "inviteFailed": "Could not send the invitation. Please try again.",
    "allProjects": "All projects",
    "invitesIntro": "Incoming invitations and decisions",
    "noInvites": "No invitations",
    "sender": "Invited by",
    "status": "Status",
    "wait": "Please wait…",
    "pending": "Pending",
    "accepted": "Accepted",
    "declined": "Declined",
    "expired": "Expired",
    "restoring": "Restoring session…",
    "codeSent": "A code was sent to your email.",
    "sixDigits": "Enter exactly six digits.",
    "welcome": "Welcome back",
    "createAccount": "Create account",
    "enterCode": "Enter the email code",
    "username": "Username",
    "password": "Password",
    "codeLabel": "Six-digit code",
    "signIn": "Sign in",
    "continue": "Continue",
    "verify": "Verify",
    "noAccount": "Create an account",
    "haveAccount": "Sign in instead",
    "avatarFormat": "Use a JPEG or PNG up to 5 MiB.",
    "choosePhoto": "Choose photo",
    "displayName": "Display name",
    "uploadFailed": "Could not upload the photo.",
    "requestFailed": "The request failed. Please try again.",
    "assignmentChanged": "Task assignee changed",
    "statusChanged": "Task status changed",
    "dueChanged": "Task due date changed",
    "taskUpdated": "Task updated",
    "invitation": "Project invitation"
  }
} as const;

const requestErrors: Record<Locale, Record<number, string>> = {
  ru: {400:"Проверьте введённые данные.",401:"Сессия завершена. Войдите снова.",403:"Нет доступа к этому действию.",404:"Данные не найдены или нет доступа.",409:"Данные уже изменились. Получите актуальную версию.",422:"Действие нарушает правила рабочего процесса.",429:"Слишком много попыток. Попробуйте позже.",503:"Сервис временно недоступен."},
  en: {400:"Check the entered information.",401:"Your session ended. Please sign in again.",403:"You do not have permission for this action.",404:"The data was not found or is not accessible.",409:"The data has changed. Please reload the latest version.",422:"This action is not allowed by the workflow.",429:"Too many attempts. Please try again later.",503:"The service is temporarily unavailable."},
};
export function localizedRequestError(error: unknown, locale: Locale): string {
  const status = error && typeof error === "object" && "status" in error ? Number(error.status) : 0;
  return requestErrors[locale][status] ?? chromeCopy[locale].requestFailed;
}
export function notificationText(kind: string, payload: Record<string, unknown>, locale: Locale): string {
  const copy = chromeCopy[locale];
  const action = ({assignment_changed:copy.assignmentChanged,status_changed:copy.statusChanged,due_at_changed:copy.dueChanged,invitation:copy.invitation} as Record<string,string>)[kind] ?? copy.taskUpdated;
  return action + (typeof payload.task_title === "string" ? `: ${payload.task_title}` : "");
}

// Workspace copy is kept here so locale changes also update open forms and errors.
export const boardCopy = {
  ru: {
    allProjects: 'Все проекты', projects: 'Проекты', task: 'Задача', newTask: 'Новая задача', team: 'Команда', projectHistory: 'История проекта',
    search: 'Поиск', searchTasks: 'Поиск задач', filters: 'Фильтры', searchFilters: 'Поиск и фильтры', activeFilters: 'Активные фильтры',
    status: 'Статус', type: 'Тип', priority: 'Приоритет', assignee: 'Исполнитель', allStatuses: 'Все статусы', allTypes: 'Все типы', allPriorities: 'Все приоритеты', allAssignees: 'Все исполнители',
    dueFrom: 'Срок с', dueTo: 'Срок по', reset: 'Сбросить', sorting: 'Сортировка', createdSort: 'По созданию', updatedSort: 'По обновлению', dueSort: 'По сроку', prioritySort: 'По приоритету', sortDirection: 'Порядок сортировки',
    emptyFilters: 'По применённым фильтрам задач не найдено. Уберите один из фильтров или сбросьте их все.', emptyColumn: 'Нет задач', unassigned: 'Не назначен', noDue: 'Без срока',
    moveDenied: 'Этот переход недоступен для задачи. Выберите разрешённый статус в меню карточки.', moveFailed: 'Не удалось изменить статус. Карточка возвращена на прежнее место.', moveConflict: 'Задача уже изменена. Карточка возвращена. Обновите доску и повторите действие.',
    changeStatus: 'Изменить статус', noTransitions: 'Нет доступных переходов', moveTo: 'Перенести в', saving: 'Сохраняем…', previous: 'Назад', next: 'Далее', of: 'из', noResults: '0 результатов',
    close: 'Закрыть', changes: 'Изменения', loadMore: 'Загрузить ещё', loading: 'Загрузка…', emptyActivity: 'История проекта пока пуста.', loadFailed: 'Не удалось загрузить данные. Попробуйте ещё раз.', membersFailed: 'Не удалось загрузить участников.', retry: 'Повторить',
    title: 'Название', description: 'Описание', markdownDescription: 'Описание (Markdown)', author: 'Автор', due: 'Срок', edit: 'Редактировать', remove: 'Удалить', cancel: 'Отмена', save: 'Сохранить', details: 'Детали', history: 'История задачи',
    saveFailed: 'Не удалось сохранить задачу. Введённые данные сохранены.', removeFailed: 'Не удалось удалить задачу.', deleteTask: 'Удалить задачу',
    comments: 'Комментарии', noComments: 'Пока нет комментариев.', commentActions: 'Действия с комментарием', editComment: 'Изменить', deleteComment: 'Удалить комментарий?', writeComment: 'Написать комментарий', send: 'Отправить', sending: 'Отправляем…',
    commentFailed: 'Не удалось отправить комментарий. Текст сохранён.', commentEditFailed: 'Не удалось сохранить комментарий. Текст сохранён.', commentDeleteFailed: 'Не удалось удалить комментарий.',
    invalid: 'Проверьте обязательные поля и введённые данные.', forbidden: 'У вас нет доступа к этому действию.', conflict: 'Данные уже изменились. Обновите состояние и повторите действие.', unavailable: 'Сервис временно недоступен. Введённые данные сохранены — попробуйте ещё раз.', sessionExpired: 'Сессия завершена. Войдите снова.', notFound: 'Данные не найдены или нет доступа.', tooMany: 'Слишком много попыток. Попробуйте позже.',
    changed: 'изменил(а)', object: 'объект', moved: 'перенёс(ла)',
    types: { task: 'Задача', bug: 'Ошибка', feature: 'Функция', research: 'Исследование' },
    priorities: { low: 'Низкий', medium: 'Средний', high: 'Высокий', critical: 'Критический' },
    actions: { 'task.created': 'создал(а) задачу', 'task.updated': 'изменил(а) задачу', 'task.deleted': 'удалил(а) задачу', 'task.status_changed': 'изменил(а) статус задачи', 'task.priority_changed': 'изменил(а) приоритет задачи', 'task.assignee_changed': 'изменил(а) исполнителя задачи', 'task.due_at_changed': 'изменил(а) срок задачи', 'comment.created': 'добавил(а) комментарий', 'comment.updated': 'изменил(а) комментарий', 'comment.deleted': 'удалил(а) комментарий', 'member.joined': 'добавил(а) участника', 'member.role_changed': 'изменил(а) роль участника', 'member.removed': 'удалил(а) участника', 'project.created': 'создал(а) проект', 'project.renamed': 'переименовал(а) проект', 'project.updated': 'изменил(а) проект', 'project.archived': 'архивировал(а) проект', 'project.restored': 'восстановил(а) проект', 'invite.sent': 'отправил(а) приглашение', 'invite.accepted': 'принял(а) приглашение', 'invite.declined': 'отклонил(а) приглашение' },
  },
  en: {
    allProjects: 'All projects', projects: 'Projects', task: 'Task', newTask: 'New task', team: 'Team', projectHistory: 'Project history',
    search: 'Search', searchTasks: 'Search tasks', filters: 'Filters', searchFilters: 'Search and filters', activeFilters: 'Active filters',
    status: 'Status', type: 'Type', priority: 'Priority', assignee: 'Assignee', allStatuses: 'All statuses', allTypes: 'All types', allPriorities: 'All priorities', allAssignees: 'All assignees',
    dueFrom: 'Due from', dueTo: 'Due through', reset: 'Reset', sorting: 'Sort by', createdSort: 'Created date', updatedSort: 'Updated date', dueSort: 'Due date', prioritySort: 'Priority', sortDirection: 'Sort direction',
    emptyFilters: 'No tasks match these filters. Remove a filter or reset them all.', emptyColumn: 'No tasks', unassigned: 'Unassigned', noDue: 'No due date',
    moveDenied: 'This transition is unavailable. Choose an allowed status from the card menu.', moveFailed: 'Could not change the status. The card was returned to its original position.', moveConflict: 'This task has already changed. The card was returned. Refresh the board and try again.',
    changeStatus: 'Change status', noTransitions: 'No available transitions', moveTo: 'Move to', saving: 'Saving…', previous: 'Previous', next: 'Next', of: 'of', noResults: '0 results',
    close: 'Close', changes: 'Changes', loadMore: 'Load more', loading: 'Loading…', emptyActivity: 'No project activity yet.', loadFailed: 'Could not load data. Please try again.', membersFailed: 'Could not load members.', retry: 'Retry',
    title: 'Title', description: 'Description', markdownDescription: 'Description (Markdown)', author: 'Author', due: 'Due date', edit: 'Edit', remove: 'Delete', cancel: 'Cancel', save: 'Save', details: 'Details', history: 'Task history',
    saveFailed: 'Could not save the task. Your input has been preserved.', removeFailed: 'Could not delete the task.', deleteTask: 'Delete task',
    comments: 'Comments', noComments: 'No comments yet.', commentActions: 'Comment actions', editComment: 'Edit', deleteComment: 'Delete this comment?', writeComment: 'Write a comment', send: 'Send', sending: 'Sending…',
    commentFailed: 'Could not send the comment. Your text has been preserved.', commentEditFailed: 'Could not save the comment. Your text has been preserved.', commentDeleteFailed: 'Could not delete the comment.',
    invalid: 'Check the required fields and your input.', forbidden: 'You do not have permission for this action.', conflict: 'The data has already changed. Refresh and try again.', unavailable: 'The service is temporarily unavailable. Your input has been preserved — please try again.', sessionExpired: 'Your session has expired. Sign in again.', notFound: 'Data was not found or is not accessible.', tooMany: 'Too many attempts. Please try again later.',
    changed: 'updated', object: 'object', moved: 'moved',
    types: { task: 'Task', bug: 'Bug', feature: 'Feature', research: 'Research' },
    priorities: { low: 'Low', medium: 'Medium', high: 'High', critical: 'Critical' },
    actions: { 'task.created': 'created task', 'task.updated': 'updated task', 'task.deleted': 'deleted task', 'task.status_changed': 'changed task status', 'task.priority_changed': 'changed task priority', 'task.assignee_changed': 'changed task assignee', 'task.due_at_changed': 'changed task due date', 'comment.created': 'added comment', 'comment.updated': 'edited comment', 'comment.deleted': 'deleted comment', 'member.joined': 'added member', 'member.role_changed': 'changed member role', 'member.removed': 'removed member', 'project.created': 'created project', 'project.renamed': 'renamed project', 'project.updated': 'updated project', 'project.archived': 'archived project', 'project.restored': 'restored project', 'invite.sent': 'sent invitation', 'invite.accepted': 'accepted invitation', 'invite.declined': 'declined invitation' },
  },
} as const;
export type BoardErrorKey = 'loadFailed' | 'membersFailed' | 'saveFailed' | 'removeFailed' | 'commentFailed' | 'commentEditFailed' | 'commentDeleteFailed';
export function boardErrorText(error: unknown, fallback: BoardErrorKey, locale: Locale) {
  const copy = boardCopy[locale];
  const status = error && typeof error === 'object' && 'status' in error ? Number(error.status) : 0;
  if (status === 400 || status === 422) return copy.invalid;
  if (status === 401) return copy.sessionExpired;
  if (status === 403) return copy.forbidden;
  if (status === 404) return copy.notFound;
  if (status === 409) return copy.conflict;
  if (status === 429) return copy.tooMany;
  if (status >= 500) return copy.unavailable;
  return copy[fallback];
}

export const activityCopy = {
  ru: {
    fields: {title:'Название',description:'Описание',goal:'Цель',status:'Статус',type:'Тип',priority:'Приоритет',role:'Роль',display_name:'Имя',username:'Имя пользователя',user_id:'Участник',invitee_id:'Приглашённый участник',invitee_name:'Имя приглашённого',assignee_id:'Исполнитель',assignee_name:'Имя исполнителя',due_at:'Срок',created_at:'Создано',updated_at:'Обновлено',completed_at:'Завершено',expires_at:'Истекает',body:'Комментарий',changes:'Изменения',changed:'Изменено',deleted:'Удалено'},
    statuses: {active:'Активен',archived:'В архиве',pending:'Ожидает ответа',accepted:'Принято',declined:'Отклонено',expired:'Истекло'},
    yes:'Да',no:'Нет',changed:'Изменено',details:'Данные',
  },
  en: {
    fields: {title:'Title',description:'Description',goal:'Goal',status:'Status',type:'Type',priority:'Priority',role:'Role',display_name:'Name',username:'Username',user_id:'Member',invitee_id:'Invited member',invitee_name:'Invitee name',assignee_id:'Assignee',assignee_name:'Assignee name',due_at:'Due date',created_at:'Created',updated_at:'Updated',completed_at:'Completed',expires_at:'Expires',body:'Comment',changes:'Changes',changed:'Changed',deleted:'Deleted'},
    statuses: {active:'Active',archived:'Archived',pending:'Pending',accepted:'Accepted',declined:'Declined',expired:'Expired'},
    yes:'Yes',no:'No',changed:'Changed',details:'Details',
  },
} as const;
