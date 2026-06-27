import { Navigate, Route, Routes } from "react-router-dom";
import DashboardPage from "./app/dashboard/page";
import { IssueDetailLayout } from "./components/layout/issue";
import ConversationRoute from "./app/issues/[id]/conversations/page";
import IssueDetailRoute from "./app/issues/[id]/page";
import IssuesPage from "./app/issues/page";
import IssuesTablePage from "./app/issues/table/page";
import ProjectSettingsPage from "./app/projects/[projectKey]/settings/page";
import SettingsPage from "./app/settings/page";
import { DefaultLayout } from "./components/layout/default";
import { Layout } from "./components/layout";
import { ProjectLayout } from "./components/layout/project";
import { TabsProvider, type TabPageLink } from "./context/tabs";

const projectHeaderPages = [
  { key: "issues", href: "/issues", titleKey: "header.board" },
  { key: "table", href: "/issues/table", titleKey: "issues.table.tableTab" },
] satisfies readonly TabPageLink[];

const projectDetailHeaderPages = [
  { key: "issues", href: "/projects/:projectKey/issues", titleKey: "header.board" },
  { key: "table", href: "/projects/:projectKey/table", titleKey: "issues.table.tableTab" },
  { key: "settings", href: "/projects/:projectKey/settings", titleKey: "header.settings" },
] satisfies readonly TabPageLink[];

const issueHeaderPages = [
  { key: "detail", href: "/issues/:id", titleKey: "issues.detailPage.detailTab" },
  {
    key: "conversations",
    href: "/issues/:id/conversations",
    titleKey: "issues.detailPage.conversationTab",
  },
] satisfies readonly TabPageLink[];

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route
          path="/issues"
          element={
            <TabsProvider activeKey="issues" pages={projectHeaderPages}>
              <ProjectLayout>
                <IssuesPage />
              </ProjectLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/issues/table"
          element={
            <TabsProvider activeKey="table" pages={projectHeaderPages}>
              <ProjectLayout>
                <IssuesTablePage />
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
        <Route
          path="/issues/:id"
          element={
            <TabsProvider activeKey="detail" pages={issueHeaderPages}>
              <IssueDetailLayout>
                <IssueDetailRoute />
              </IssueDetailLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/issues/:id/conversations"
          element={
            <TabsProvider activeKey="conversations" pages={issueHeaderPages}>
              <IssueDetailLayout>
                <ConversationRoute />
              </IssueDetailLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/issues/:id/runs/:runId/conversations"
          element={
            <TabsProvider activeKey="conversations" pages={issueHeaderPages}>
              <IssueDetailLayout>
                <ConversationRoute />
              </IssueDetailLayout>
            </TabsProvider>
          }
        />
        <Route
          path="/dashboard"
          element={
            <TabsProvider>
              <DefaultLayout>
                <DashboardPage />
              </DefaultLayout>
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
        <Route path="*" element={<Navigate to="/issues" replace />} />
      </Routes>
    </Layout>
  );
}
