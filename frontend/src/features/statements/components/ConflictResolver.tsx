import { useState } from 'react';
import type { Statement, ConflictResolution } from '../types';
import { useResolveConflict } from '../hooks/useStatements';

interface ConflictResolverProps {
  statement: Statement;
  onClose: () => void;
  onResolved?: (statement: Statement) => void;
}

export function ConflictResolver({
  statement,
  onClose,
  onResolved,
}: ConflictResolverProps) {
  const [resolution, setResolution] = useState<ConflictResolution>('keep_local');
  const [mergedContent, setMergedContent] = useState(
    statement.local_content || ''
  );

  const resolveMutation = useResolveConflict();

  const handleResolve = async () => {
    try {
      const result = await resolveMutation.mutateAsync({
        id: statement.id,
        request: {
          resolution,
          merged_content: resolution === 'merge' ? mergedContent : undefined,
        },
      });
      onResolved?.(result);
      onClose();
    } catch (error) {
      console.error('Failed to resolve conflict:', error);
    }
  };

  return (
    <div className="conflict-resolver">
      <div className="conflict-resolver-header">
        <h3>Resolve Conflict</h3>
        <button
          type="button"
          className="conflict-resolver-close"
          onClick={onClose}
          aria-label="Close"
        >
          &times;
        </button>
      </div>

      <div className="alert alert-warning">
        This statement was modified both locally and remotely. Choose how to
        resolve this conflict.
      </div>

      <div className="conflict-resolver-comparison">
        <div className="conflict-resolver-version">
          <h4>Remote Version</h4>
          <div className="conflict-resolver-meta">
            {statement.remote_updated_at && (
              <span>
                Updated:{' '}
                {new Date(statement.remote_updated_at).toLocaleString()}
              </span>
            )}
          </div>
          <pre className="conflict-resolver-content">
            {statement.remote_content || 'No content'}
          </pre>
        </div>

        <div className="conflict-resolver-version">
          <h4>Local Version</h4>
          <div className="conflict-resolver-meta">
            {statement.modified_at && (
              <span>
                Modified: {new Date(statement.modified_at).toLocaleString()}
              </span>
            )}
          </div>
          <pre className="conflict-resolver-content">
            {statement.local_content || 'No content'}
          </pre>
        </div>
      </div>

      <div className="conflict-resolver-options">
        <h4>Resolution Strategy</h4>

        <label className="conflict-resolver-option">
          <input
            type="radio"
            name="resolution"
            value="keep_local"
            checked={resolution === 'keep_local'}
            onChange={(e) =>
              setResolution(e.target.value as ConflictResolution)
            }
          />
          <div className="conflict-resolver-option-info">
            <strong>Keep Local</strong>
            <p>
              Discard remote changes and keep your local modifications. Your
              changes will be pushed to ServiceNow on the next sync.
            </p>
          </div>
        </label>

        <label className="conflict-resolver-option">
          <input
            type="radio"
            name="resolution"
            value="keep_remote"
            checked={resolution === 'keep_remote'}
            onChange={(e) =>
              setResolution(e.target.value as ConflictResolution)
            }
          />
          <div className="conflict-resolver-option-info">
            <strong>Keep Remote</strong>
            <p>
              Discard your local changes and accept the remote version from
              ServiceNow.
            </p>
          </div>
        </label>

        <label className="conflict-resolver-option">
          <input
            type="radio"
            name="resolution"
            value="merge"
            checked={resolution === 'merge'}
            onChange={(e) =>
              setResolution(e.target.value as ConflictResolution)
            }
          />
          <div className="conflict-resolver-option-info">
            <strong>Manual Merge</strong>
            <p>
              Manually combine both versions into a merged result below.
            </p>
          </div>
        </label>
      </div>

      {resolution === 'merge' && (
        <div className="conflict-resolver-merge">
          <label htmlFor="merged-content">Merged Content</label>
          <textarea
            id="merged-content"
            value={mergedContent}
            onChange={(e) => setMergedContent(e.target.value)}
            rows={10}
            placeholder="Enter the merged content..."
          />
        </div>
      )}

      {resolveMutation.isError && (
        <div className="alert alert-error">
          Failed to resolve conflict. Please try again.
        </div>
      )}

      <div className="conflict-resolver-actions">
        <button
          type="button"
          className="btn btn-secondary"
          onClick={onClose}
        >
          Cancel
        </button>
        <button
          type="button"
          className="btn btn-primary"
          onClick={handleResolve}
          disabled={resolveMutation.isPending}
        >
          {resolveMutation.isPending ? 'Resolving...' : 'Resolve Conflict'}
        </button>
      </div>
    </div>
  );
}

export default ConflictResolver;
