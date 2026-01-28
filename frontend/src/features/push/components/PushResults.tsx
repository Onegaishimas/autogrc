import type { PushJob } from '../types';
import { getPushStatusInfo } from '../types';

interface PushResultsProps {
  job: PushJob;
  onDone: () => void;
  onRetry?: () => void;
}

export function PushResults({ job, onDone, onRetry }: PushResultsProps) {
  const statusInfo = getPushStatusInfo(job.status);
  const hasFailures = job.failed > 0;
  const allFailed = job.succeeded === 0 && job.failed > 0;

  return (
    <div className="push-results">
      <div className="push-results-header">
        <h3>Push Complete</h3>
        <span
          className="push-results-badge"
          style={{
            color: statusInfo.color,
            backgroundColor: statusInfo.bgColor,
          }}
        >
          {statusInfo.label}
        </span>
      </div>

      <div className="push-results-summary">
        <div className="push-results-stat">
          <span className="stat-value">{job.total_count}</span>
          <span className="stat-label">Total</span>
        </div>
        <div className="push-results-stat success">
          <span className="stat-value">{job.succeeded}</span>
          <span className="stat-label">Succeeded</span>
        </div>
        {hasFailures && (
          <div className="push-results-stat error">
            <span className="stat-value">{job.failed}</span>
            <span className="stat-label">Failed</span>
          </div>
        )}
      </div>

      {hasFailures && (
        <div className="push-results-errors">
          <h4>Failed Statements</h4>
          <ul className="push-results-error-list">
            {job.results
              .filter((r) => !r.success)
              .map((r) => (
                <li key={r.statement_id}>
                  <span className="error-id">{r.statement_id.slice(0, 8)}...</span>
                  <span className="error-message">{r.error || 'Unknown error'}</span>
                </li>
              ))}
          </ul>
        </div>
      )}

      {!hasFailures && (
        <div className="alert alert-success">
          All statements were successfully pushed to ServiceNow.
        </div>
      )}

      {job.completed_at && (
        <p className="push-results-time">
          Completed at {new Date(job.completed_at).toLocaleString()}
        </p>
      )}

      <div className="push-results-actions">
        {allFailed && onRetry && (
          <button
            type="button"
            className="btn btn-secondary"
            onClick={onRetry}
          >
            Retry
          </button>
        )}
        <button type="button" className="btn btn-primary" onClick={onDone}>
          Done
        </button>
      </div>
    </div>
  );
}

export default PushResults;
