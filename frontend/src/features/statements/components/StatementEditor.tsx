import { useState, useEffect } from 'react';
import type { Statement } from '../types';
import { getSyncStatusInfo } from '../types';
import { useUpdateStatement, useRevertStatement } from '../hooks/useStatements';

interface StatementEditorProps {
  statement: Statement;
  onClose: () => void;
  onSaved?: (statement: Statement) => void;
}

export function StatementEditor({
  statement,
  onClose,
  onSaved,
}: StatementEditorProps) {
  const [content, setContent] = useState(statement.effective_content || '');
  const [isDirty, setIsDirty] = useState(false);

  const updateMutation = useUpdateStatement();
  const revertMutation = useRevertStatement();

  const statusInfo = getSyncStatusInfo(statement.sync_status);

  useEffect(() => {
    setContent(statement.effective_content || '');
    setIsDirty(false);
  }, [statement]);

  const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setContent(e.target.value);
    setIsDirty(e.target.value !== statement.effective_content);
  };

  const handleSave = async () => {
    try {
      const result = await updateMutation.mutateAsync({
        id: statement.id,
        request: { local_content: content },
      });
      setIsDirty(false);
      onSaved?.(result);
    } catch (error) {
      console.error('Failed to save statement:', error);
    }
  };

  const handleRevert = async () => {
    if (!confirm('Revert to remote content? Local changes will be lost.')) {
      return;
    }
    try {
      const result = await revertMutation.mutateAsync(statement.id);
      setContent(result.effective_content || '');
      setIsDirty(false);
      onSaved?.(result);
    } catch (error) {
      console.error('Failed to revert statement:', error);
    }
  };

  const handleClose = () => {
    if (isDirty) {
      if (!confirm('You have unsaved changes. Discard them?')) {
        return;
      }
    }
    onClose();
  };

  return (
    <div className="statement-editor">
      <div className="statement-editor-header">
        <div className="statement-editor-title">
          <h3>Edit Statement</h3>
          <span
            className="statement-editor-badge"
            style={{
              color: statusInfo.color,
              backgroundColor: statusInfo.bgColor,
            }}
          >
            {statusInfo.label}
          </span>
        </div>
        <button
          type="button"
          className="statement-editor-close"
          onClick={handleClose}
          aria-label="Close"
        >
          &times;
        </button>
      </div>

      <div className="statement-editor-info">
        <div className="statement-editor-info-row">
          <span className="label">Type:</span>
          <span>{statement.statement_type}</span>
        </div>
        {statement.remote_updated_at && (
          <div className="statement-editor-info-row">
            <span className="label">Remote updated:</span>
            <span>
              {new Date(statement.remote_updated_at).toLocaleString()}
            </span>
          </div>
        )}
      </div>

      {statement.sync_status === 'conflict' && (
        <div className="alert alert-warning">
          This statement has a conflict. The remote content has changed since
          your last edit. Please resolve the conflict or use the conflict
          resolver.
        </div>
      )}

      <div className="statement-editor-content">
        <label htmlFor="statement-content">Content</label>
        <textarea
          id="statement-content"
          value={content}
          onChange={handleContentChange}
          rows={12}
          placeholder="Enter statement content..."
        />
      </div>

      {statement.is_modified && statement.remote_content && (
        <details className="statement-editor-remote">
          <summary>View remote content</summary>
          <pre>{statement.remote_content}</pre>
        </details>
      )}

      {(updateMutation.isError || revertMutation.isError) && (
        <div className="alert alert-error">
          Failed to save changes. Please try again.
        </div>
      )}

      <div className="statement-editor-actions">
        <button
          type="button"
          className="btn btn-secondary"
          onClick={handleClose}
        >
          Cancel
        </button>
        {statement.is_modified && (
          <button
            type="button"
            className="btn btn-warning"
            onClick={handleRevert}
            disabled={revertMutation.isPending}
          >
            {revertMutation.isPending ? 'Reverting...' : 'Revert to Remote'}
          </button>
        )}
        <button
          type="button"
          className="btn btn-primary"
          onClick={handleSave}
          disabled={!isDirty || updateMutation.isPending}
        >
          {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </div>
  );
}

export default StatementEditor;
