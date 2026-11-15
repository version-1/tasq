# Project document artifacts

Markdown artifacts are owned by one project. `document_artifact_links` records optional polymorphic references; the first supported target is `issue`. An artifact can have no links or be linked to issues in any project. The owner-project route is the canonical route for reads and mutations, so a link does not become an ownership or authorization grant. The [API design](api.md#document-artifacts) describes the limited compatibility issue routes.

Deleting an issue cascades only its links. Deleting an artifact is rejected while links exist. Deleting an owner project is rejected when one of its artifacts is linked to an issue in another project. The database trigger enforces this invariant at the delete boundary. The store holds the tracker write transaction while it checks the invariant, cleans up the other store's runtime data, and deletes the project, so a concurrent link cannot leave only runtime data deleted.

The migration converts each existing issue-owned document into a project-owned artifact. It preserves the ID, body, type, title, timestamps, and creates a link back to its original issue. Its down migration restores the old table only if every artifact still has exactly one same-project issue link. It otherwise rejects the rollback before changing data. Operators must use a stopped-service SQLite backup to recover from a valid state that the older representation cannot express.

This service currently has no authentication or authorization middleware. Owner-project routing preserves the data boundary for a future policy layer but does not itself enforce user permissions.
