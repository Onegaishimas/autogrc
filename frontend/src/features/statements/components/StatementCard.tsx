import type { Statement } from '../types';
import { getSyncStatusInfo } from '../types';

interface StatementCardProps {
  statement: Statement;
  onEdit?: (statement: Statement) => void;
  onResolve?: (statement: Statement) => void;
  showActions?: boolean;
}

export function StatementCard({
  statement,
  onEdit,
  onResolve,
  showActions = true,
}: StatementCardProps) {
  const statusInfo = getSyncStatusInfo(statement.sync_status);

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const truncateContent = (content: string, maxLength = 200) => {
    if (content.length <= maxLength) return content;
    return content.slice(0, maxLength) + '...';
  };

  return (
    <div className={`statement-card ${statement.sync_status}`}>
      <div className="statement-card-header">
        <div className="statement-card-type">{statement.statement_type}</div>
        <span
          className="statement-card-badge"
          style={{
            color: statusInfo.color,
            backgroundColor: statusInfo.bgColor,
          }}
        >
          {statusInfo.label}
        </span>
      </div>

      <div className="statement-card-content">
        <p>{truncateContent(statement.effective_content || 'No content')}</p>
      </div>

      <div className="statement-card-meta">
        {statement.is_modified && statement.modified_at && (
          <div className="statement-card-meta-item">
            <span className="statement-card-meta-label">Modified:</span>
            <span>{formatDate(statement.modified_at)}</span>
          </div>
        )}
        {statement.last_pull_at && (
          <div className="statement-card-meta-item">
            <span className="statement-card-meta-label">Last pulled:</span>
            <span>{formatDate(statement.last_pull_at)}</span>
          </div>
        )}
      </div>

      {showActions && (
        <div className="statement-card-actions">
          {statement.sync_status === 'conflict' && onResolve && (
            <button
              type="button"
              className="btn btn-sm btn-warning"
              onClick={() => onResolve(statement)}
            >
              Resolve Conflict
            </button>
          )}
          {onEdit && (
            <button
              type="button"
              className="btn btn-sm btn-secondary"
              onClick={() => onEdit(statement)}
            >
              Edit
            </button>
          )}
        </div>
      )}
    </div>
  );
}

export default StatementCard;
