# プロジェクト文書 artifact

Markdown artifact は一つの project が所有します。`document_artifact_links` は任意の polymorphic な参照を保持し、初期対応する対象は `issue` です。artifact は関連なしでも作成でき、任意の project の issue に複数関連できます。所有 project の route が取得と更新の正規 route であるため、関連によって所有権や認可を与えません。限定的な互換性のある課題 route は [API 設計](api.ja.md#document-artifact)を参照してください。

課題を削除すると関連だけが cascade されます。関連が残る artifact の削除は拒否します。所有プロジェクトの削除は、別プロジェクトの課題が関連する artifact を含む場合に拒否します。database trigger が削除境界でこの不変条件を保証します。Store は Tracker の書き込みトランザクションを保持したまま不変条件を確認し、別ストアの実行時データをクリーンアップしてからプロジェクトを削除します。そのため、並行した関連追加によって実行時データだけが削除されることはありません。

migration は既存の課題所有文書をプロジェクト所有 artifact に変換します。ID、本文、種別、タイトル、時刻を保持し、元の課題への関連を作成します。down migration は各 artifact が同一プロジェクトの課題への関連をちょうど 1 件だけ持つ場合に限って旧テーブルへ戻します。それ以外では、データを変更する前に rollback を拒否します。古い形式で表せない有効な状態から復旧する場合、運用者はサービスを停止した SQLite バックアップを使用します。

現サービスには認証・認可 middleware がありません。所有 project route は将来の policy 層のデータ境界を保ちますが、利用者権限を単独で強制するものではありません。
