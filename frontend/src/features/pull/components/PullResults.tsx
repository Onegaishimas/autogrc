import type { PullJob } from '../types';

interface PullResultsProps {
  job: PullJob;
  onDone: () => void;
  onRetry?: () => void;
}

export function PullResults({ job, onDone, onRetry }: PullResultsProps) {
  const isSuccess = job.status === 'completed';
  const isFailed = job.status === 'failed';
  const isCancelled = job.status === 'cancelled';

  const { progress } = job;
  const hasErrors = progress.errors && progress.errors.length > 0;

  return (
    <div className={`pull-results ${job.status}`}>
      <div className="pull-results-header">
        {isSuccess && (
          <>
            <span className="result-icon success">✅</span>
            <h3>Pull Completed Successfully</h3>
          </>
        )}
        {isFailed && (
          <>
            <span className="result-icon failed">❌</span>
            <h3>Pull Failed</h3>
          </>
        )}
        {isCancelled && (
          <>
            <span className="result-icon cancelled">🚫</span>
            <h3>Pull Cancelled</h3>
          </>
        )}
      </div>

      <div className="pull-results-summary">
        <h4>Summary</h4>
        <dl className="results-list">
          <div className="result-item">
            <dt>Systems Processed</dt>
            <dd>
              {progress.completed_systems} of {progress.total_systems}
            </dd>
          </div>
          <div className="result-item">
            <dt>Controls Imported</dt>
            <dd>{progress.completed_controls}</dd>
          </div>
          <div className="result-item">
            <dt>Statements Imported</dt>
            <dd>{progress.completed_statements}</dd>
          </div>
        </dl>
      </div>

      {job.error && (
        <div className="pull-results-error alert alert-error">
          <strong>Error:</strong> {job.error}
        </div>
      )}

      {hasErrors && (
        <div className="pull-results-warnings alert alert-warning">
          <h4>Warnings ({progress.errors!.length})</h4>
          <ul>
            {progress.errors!.map((error, index) => (
              <li key={index}>{error}</li>
            ))}
          </ul>
        </div>
      )}

      {isSuccess && (
        <div className="pull-results-next">
          <p>
            Your systems, controls, and implementation statements have been
            imported. You can now view and edit the statements.
          </p>
        </div>
      )}

      {isFailed && onRetry && (
        <div className="pull-results-next">
          <p>
            The pull operation failed. You can try again or contact support if
            the problem persists.
          </p>
        </div>
      )}

      <div className="pull-results-actions">
        {isFailed && onRetry && (
          <button
            type="button"
            className="btn btn-primary"
            onClick={onRetry}
          >
            Try Again
          </button>
        )}
        <button
          type="button"
          className={isFailed ? 'btn btn-secondary' : 'btn btn-primary'}
          onClick={onDone}
        >
          {isSuccess ? 'View Systems' : 'Done'}
        </button>
      </div>
    </div>
  );
}

export default PullResults;
