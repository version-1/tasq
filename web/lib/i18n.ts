"use client";

import i18n from "i18next";
import { initReactI18next } from "react-i18next";

export const supportedLanguages = ["ja", "en"] as const;

export type SupportedLanguage = (typeof supportedLanguages)[number];

const resources = {
  ja: {
    translation: {
      agents: {
        empty: "まだエージェント実行はありません",
        issueLabel: "Issue #{{id}}",
        title: "エージェント実行",
        workspacePending: "workspace 未確定",
      },
      common: {
        pending: "未確定",
      },
      header: {
        addTask: "タスクを追加",
        agentRuns: "エージェント実行",
        agents: "エージェント",
        board: "ボード",
        breadcrumb: "パンくずリスト",
        commandKey: "⌘ K",
        filter: "▽ フィルター",
        issueCount: "{{count}} issues",
        loading: "loading",
        moreProjectActions: "プロジェクト操作",
        notifications: "通知",
        productWebsite: "Product Website",
        project: "プロジェクト",
        projectName: "Product Website",
        searchPlaceholder: "検索...",
        settings: "設定",
        sort: "↕ ソート",
        taskOptions: "タスクオプション",
        trackedRunCount: "{{count}} tracked runs",
        trackedRunLoading: "...",
        view: "▦ 表示",
        views: "ビュー",
      },
      issues: {
        assignee: "担当者",
        attempt: "試行回数",
        board: {
          addTask: "タスクを追加",
          columnActions: "カラム操作",
          draft: "Draft",
          inProgress: "In Progress",
          inReview: "In Review",
          todo: "To Do",
        },
        detail: {
          issueStatus: "Issue ステータス",
          priority: "優先度",
          runStatus: "実行ステータス",
          workspace: "Workspace",
        },
        moveLabel: "{{title}} を移動",
        noDescription: "説明なし",
        noIssueSelected: "Issue が選択されていません",
        noIssues: "Issue はありません",
        noRun: "実行なし",
        unassigned: "未割り当て",
      },
      layout: {
        apiUnavailable: "API を利用できません",
        failedToLoadSummary: "summary の読み込みに失敗しました",
        failedToUpdateIssue: "issue の更新に失敗しました",
        loading: "読み込み中",
        useLayoutDataError: "useLayoutData は Layout 内で使用してください",
      },
      priorities: {
        high: "high",
        low: "low",
        normal: "normal",
        urgent: "urgent",
      },
      settings: {
        apiOrigin: "API origin",
        language: "表示言語",
        languageEnglish: "English",
        languageJapanese: "日本語",
        lastSummary: "Last summary",
        refreshInterval: "Refresh interval",
        refreshOneSecond: "1 second",
        refreshTenSeconds: "10 seconds",
        refreshThreeSeconds: "3 seconds",
        refreshFiveSeconds: "5 seconds",
        title: "Web UI Settings",
      },
      sidebar: {
        accountMenu: "アカウントメニュー",
        addProject: "プロジェクトを追加",
        allProjects: "すべてのプロジェクト",
        apiBackend: "API Backend",
        archive: "アーカイブ",
        board: "ボード",
        calendar: "カレンダー",
        home: "tasq ホーム",
        marketing: "Marketing",
        members: "メンバー",
        mobileApp: "Mobile App",
        myTasks: "マイタスク",
        primaryNavigation: "メインナビゲーション",
        productWebsite: "Product Website",
        projects: "プロジェクト",
        reports: "レポート",
        settings: "設定",
        userEmail: "yuki@tasq.app",
        userName: "Yuki K.",
        workflow: "ワークフロー",
      },
      runStatuses: {
        cancelled: "キャンセル",
        failed: "失敗",
        queued: "待機中",
        running: "実行中",
        starting: "開始中",
        succeeded: "成功",
        waiting_for_input: "入力待ち",
      },
      statuses: {
        backlog: "Backlog",
        blocked: "Blocked",
        done: "Done",
        failed: "Failed",
        in_progress: "In Progress",
        ready: "Ready",
        review: "Review",
      },
    },
  },
  en: {
    translation: {
      agents: {
        empty: "No orchestrator runs yet",
        issueLabel: "Issue #{{id}}",
        title: "Agent Runs",
        workspacePending: "workspace pending",
      },
      common: {
        pending: "pending",
      },
      header: {
        addTask: "Add task",
        agentRuns: "Agent Runs",
        agents: "Agents",
        board: "Board",
        breadcrumb: "Breadcrumb",
        commandKey: "⌘ K",
        filter: "▽ Filter",
        issueCount: "{{count}} issues",
        loading: "loading",
        moreProjectActions: "More project actions",
        notifications: "Notifications",
        productWebsite: "Product Website",
        project: "Project",
        projectName: "Product Website",
        searchPlaceholder: "Search...",
        settings: "Settings",
        sort: "↕ Sort",
        taskOptions: "Task options",
        trackedRunCount: "{{count}} tracked runs",
        trackedRunLoading: "...",
        view: "▦ View",
        views: "Views",
      },
      issues: {
        assignee: "Assignee",
        attempt: "Attempt",
        board: {
          addTask: "Add task",
          columnActions: "Column actions",
          draft: "Draft",
          inProgress: "In Progress",
          inReview: "In Review",
          todo: "To Do",
        },
        detail: {
          issueStatus: "Issue Status",
          priority: "Priority",
          runStatus: "Run Status",
          workspace: "Workspace",
        },
        moveLabel: "Move {{title}}",
        noDescription: "No description",
        noIssueSelected: "No issue selected",
        noIssues: "No issues",
        noRun: "no run",
        unassigned: "unassigned",
      },
      layout: {
        apiUnavailable: "API unavailable",
        failedToLoadSummary: "failed to load summary",
        failedToUpdateIssue: "failed to update issue",
        loading: "Loading",
        useLayoutDataError: "useLayoutData must be used inside Layout",
      },
      priorities: {
        high: "high",
        low: "low",
        normal: "normal",
        urgent: "urgent",
      },
      settings: {
        apiOrigin: "API origin",
        language: "Display language",
        languageEnglish: "English",
        languageJapanese: "Japanese",
        lastSummary: "Last summary",
        refreshInterval: "Refresh interval",
        refreshOneSecond: "1 second",
        refreshTenSeconds: "10 seconds",
        refreshThreeSeconds: "3 seconds",
        refreshFiveSeconds: "5 seconds",
        title: "Web UI Settings",
      },
      sidebar: {
        accountMenu: "Account menu",
        addProject: "Add project",
        allProjects: "All projects",
        apiBackend: "API Backend",
        archive: "Archive",
        board: "Board",
        calendar: "Calendar",
        home: "tasq home",
        marketing: "Marketing",
        members: "Members",
        mobileApp: "Mobile App",
        myTasks: "My tasks",
        primaryNavigation: "Primary navigation",
        productWebsite: "Product Website",
        projects: "Projects",
        reports: "Reports",
        settings: "Settings",
        userEmail: "yuki@tasq.app",
        userName: "Yuki K.",
        workflow: "Workflow",
      },
      runStatuses: {
        cancelled: "cancelled",
        failed: "failed",
        queued: "queued",
        running: "running",
        starting: "starting",
        succeeded: "succeeded",
        waiting_for_input: "waiting for input",
      },
      statuses: {
        backlog: "Backlog",
        blocked: "Blocked",
        done: "Done",
        failed: "Failed",
        in_progress: "In Progress",
        ready: "Ready",
        review: "Review",
      },
    },
  },
} as const;

function initialLanguage(): SupportedLanguage {
  if (typeof window === "undefined") {
    return "ja";
  }

  const stored = window.localStorage.getItem("tasq.language");
  if (isSupportedLanguage(stored)) {
    return stored;
  }

  return navigator.language.toLowerCase().startsWith("en") ? "en" : "ja";
}

export function isSupportedLanguage(value: string | null): value is SupportedLanguage {
  return value === "ja" || value === "en";
}

if (!i18n.isInitialized) {
  void i18n.use(initReactI18next).init({
    fallbackLng: "ja",
    interpolation: {
      escapeValue: false,
    },
    lng: initialLanguage(),
    react: {
      useSuspense: false,
    },
    resources,
    supportedLngs: supportedLanguages,
  });
}

export { i18n };
