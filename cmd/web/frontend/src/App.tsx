import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import DashboardPage from "./app/dashboard/page";
import DashboardStatsPage from "./app/dashboard/stats/page";
import DashboardTablePage from "./app/dashboard/table/page";
import { NotFoundRoute } from "./app/not-found/page";
import ConversationRoute from "./app/issues/[id]/conversations/page";
import IssueDetailRoute from "./app/issues/[id]/page";
import IssuesPage from "./app/issues/page";
import IssuesTablePage from "./app/issues/table/page";
import ProjectSettingsPage from "./app/projects/[projectKey]/settings/page";
import SettingsPage from "./app/settings/page";
import { IssueDetailLayout } from "./components/layout/issue";
import { DefaultLayout } from "./components/layout/default";
import { Layout } from "./components/layout";
import { ProjectLayout } from "./components/layout/project";
import { TabsProvider, type TabPageLink } from "./context/tabs";

const dashboardHeaderPages = [
  { key: "board", href: "/dashboard", titleKey: "header.board" },
  { key: "table", href: "/dashboard/table", titleKey: "issues.table.tableTab" },
  { key: "stats", href: "/dashboard/stats", titleKey: "dashboard.statsTab" },
] satisfies readonly TabPageLink[];

const projectDetailHeaderPages = [
  { key: "issues", href: "/projects/:projectKey/issues", titleKey: "header.board" },
  { key: "table", href: "/projects/:projectKey/table", titleKey: "issues.table.tableTab" },
  { key: "settings", href: "/projects/:projectKey/settings", titleKey: "header.settings" },
] satisfies readonly TabPageLink[];

const issueHeaderPages = [
  { key: "details", href: "/issues/:id", titleKey: "issues.detailPage.detailTab" },
  { key: "comments", href: "/issues/:id?tab=comments", titleKey: "issues.detailPage.activityTab" },
  {
    key: "conversation",
    href: "/issues/:id?tab=conversation",
    titleKey: "issues.detailPage.conversationTab",
  },
] satisfies readonly TabPageLink[];

export function App() {
  const { search } = useLocation();
  const issueActiveTab = issueDetailActiveTab(search);

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route
          path="/dashboard"
          element={
            <TabsProvider activeKey="board" pages={dashboardHeaderPages}>
              <ProjectLayout>
                <DashboardPage />
              </ProjectLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/dashboard/table"
          element={
            <TabsProvider activeKey="table" pages={dashboardHeaderPages}>
              <ProjectLayout>
                <DashboardTablePage />
              </ProjectLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/projects/:projectKey"
          element={<Navigate to="issues" replace />}
        />
        <Route
          path="/projects/:projectKey/issues"
          element={
            <TabsProvider activeKey="issues" pages={projectDetailHeaderPages}>
              <ProjectLayout>
                <IssuesPage />
              </ProjectLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/projects/:projectKey/table"
          element={
            <TabsProvider activeKey="table" pages={projectDetailHeaderPages}>
              <ProjectLayout>
                <IssuesTablePage />
              </ProjectLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/projects/:projectKey/settings"
          element={
            <TabsProvider activeKey="settings" pages={projectDetailHeaderPages}>
              <ProjectLayout showAddTaskButton={false}>
                <ProjectSettingsPage />
              </ProjectLayout>
            </TabsProvider>
          }
        />
        <Route path="/issues" element={<NotFoundRoute />} />
        <Route path="/issues/table" element={<NotFoundRoute />} />
        <Route
          path="/issues/:id"
          element={
            <TabsProvider activeKey={issueActiveTab} pages={issueHeaderPages}>
              <IssueDetailLayout>
                <IssueDetailRoute />
              </IssueDetailLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/issues/:id/conversations"
          element={
            <TabsProvider activeKey="conversation" pages={issueHeaderPages}>
              <IssueDetailLayout>
                <ConversationRoute />
              </IssueDetailLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/issues/:id/runs/:runId/conversations"
          element={
            <TabsProvider activeKey="conversation" pages={issueHeaderPages}>
              <IssueDetailLayout>
                <ConversationRoute />
              </IssueDetailLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/dashboard/stats"
          element={
            <TabsProvider activeKey="stats" pages={dashboardHeaderPages}>
              <ProjectLayout showAddTaskButton={false}>
                <DashboardStatsPage />
              </ProjectLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/settings"
          element={
            <TabsProvider>
              <DefaultLayout>
                <SettingsPage />
              </DefaultLayout>
            </TabsProvider>
          }
        />
        <Route path="*" element={<NotFoundRoute />} />
      </Routes>
    </Layout>
  );
}

function issueDetailActiveTab(search: string): string {
  const tab = new URLSearchParams(search).get("tab");
  if (tab === "comments" || tab === "conversation") {
    return tab;
  }
  return "details";
}
