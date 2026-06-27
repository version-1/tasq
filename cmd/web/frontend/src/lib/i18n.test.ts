import { describe, expect, it } from "vitest";
import { i18n, supportedLanguages } from "./i18n";

const knownErrorCodes = [
  "attachments.content.file_not_found",
  "attachments.content.invalid_input",
  "attachments.content.not_found",
  "attachments.content.storage_unavailable",
  "attachments.create.entity_not_found",
  "attachments.create.file_required",
  "attachments.create.invalid_file",
  "attachments.create.invalid_file_type",
  "attachments.create.invalid_input",
  "attachments.create.invalid_request",
  "attachments.create.storage_error",
  "attachments.create.storage_unavailable",
  "attachments.delete.invalid_input",
  "attachments.delete.not_found",
  "attachments.delete.storage_unavailable",
  "attachments.list.invalid_input",
  "comments.create.invalid_id",
  "comments.create.invalid_input",
  "comments.create.invalid_request",
  "comments.create.issue_not_found",
  "comments.list.invalid_id",
  "comments.list.invalid_input",
  "comments.list.issue_not_found",
  "comments.update.invalid_id",
  "comments.update.invalid_input",
  "comments.update.invalid_request",
  "comments.update.not_found",
  "clipboardUnavailable",
  "conversation_query_failed",
  "invalid_issue_identifier",
  "invalid_run_id",
  "issue_not_found",
  "issue_query_failed",
  "issues.create.invalid_input",
  "issues.create.invalid_request",
  "issues.get.invalid_id",
  "issues.get.invalid_input",
  "issues.get.not_found",
  "issues.list.internal_error",
  "issues.list.invalid_limit",
  "issues.list.invalid_offset",
  "issues.list.invalid_project_id",
  "issues.list.invalid_project_ids",
  "issues.list.invalid_priorities",
  "issues.list.invalid_states",
  "issues.states.internal_error",
  "issues.states.invalid_request",
  "issues.update.invalid_id",
  "issues.update.invalid_input",
  "issues.update.invalid_request",
  "issues.update.not_found",
  "orchestrator_error",
  "projects.check.invalid_id",
  "projects.check.invalid_input",
  "projects.check.invalid_request",
  "projects.check.not_found",
  "projects.create.invalid_input",
  "projects.create.invalid_request",
  "projects.delete.invalid_id",
  "projects.delete.invalid_input",
  "projects.delete.not_found",
  "projects.get.invalid_id",
  "projects.get.invalid_input",
  "projects.get.not_found",
  "projects.list.internal_error",
  "projects.workflow.not_found",
  "projects.update.invalid_id",
  "projects.update.invalid_input",
  "projects.update.invalid_request",
  "projects.update.not_found",
  "refresh_unavailable",
  "run_not_found",
  "state_query_failed",
  "summary.get.internal_error",
] as const;

const genericSuccessActions = ["created", "deleted", "saved", "threadIDCopied", "updated"] as const;

describe("i18n toast resources", () => {
  it("has translations for known server error codes in every supported language", () => {
    for (const language of supportedLanguages) {
      for (const code of knownErrorCodes) {
        expect(i18n.exists(`toast.error.${code}`, { lng: language }), `${language}: ${code}`).toBe(
          true,
        );
      }
    }
  });

  it("has generic success action translations in every supported language", () => {
    for (const language of supportedLanguages) {
      for (const action of genericSuccessActions) {
        expect(i18n.exists(`toast.success.${action}`, { lng: language }), `${language}: ${action}`).toBe(
          true,
        );
      }
    }
  });
});
