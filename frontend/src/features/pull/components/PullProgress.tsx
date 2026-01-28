import type { PullJob, PullProgress as PullProgressType } from '../types';
import { calculatePullProgress, isPullJobActive } from '../types';

interface PullProgressProps {
  job: PullJob;
  onCancel?: () => void;
}

function getStatusIcon(status: PullJob['status']): string {
  switch (status) {
    case 'pending':
      return '⏳';
    case 'running':
      return '🔄';
    case 'completed':
      return '✅';
    case 'failed':
      return '❌';
    case 'cancelled':
      return '🚫';
    default:
      return '❓';
  }
}

function getStatusLabel(status: PullJob['status']): string {
  switch (status) {
    case 'pending':
      return 'Waiting to start...';
    case 'running':
      return 'Pulling data from ServiceNow...';
    case 'completed':
      return 'Pull completed successfully';
    case 'failed':
      return 'Pull failed';
    case 'cancelled':
      return 'Pull cancelled';
    default:
      return 'Unknown status';
  }
}

function ProgressBar({ value, max }: { value: number; max: number }) {
  const percentage = max > 0 ? Math.round((value / max) * 100) : 0;

  return (
    <div className="progress-bar-container">
      <div className="progress-bar">
        <div
          className="progress-bar-fill"
          style={{ width: `${percentage}%` }}
          role="progressbar"
          aria-valuenow={value}
          aria-valuemin={0}
          aria-valuemax={max}
        />
      </div>
      <span className="progress-bar-label">
        {value} / {max}
      </span>
    </div>
  );
}

function ProgressStats({ progress }: { progress: PullProgressType }) {
  return (
    <div className="pull-progress-stats">
      <div className="progress-stat">
        <span className="progress-stat-label">Systems</span>
        <ProgressBar
          value={progress.completed_systems}
          max={progress.total_systems}
        />
      </div>

      {progress.total_controls > 0 && (
        <div className="progress-stat">
          <span className="progress-stat-label">Controls</span>
          <ProgressBar
            value={progress.completed_controls}
            max={progress.total_controls}
          />
        </div>
      )}

      {progress.total_statements > 0 && (
        <div className="progress-stat">
          <span className="progress-stat-label">Statements</span>
          <ProgressBar
            value={progress.completed_statements}
            max={progress.total_statements}
          />
        </div>
      )}

      {progress.current_system && (
        <div className="progress-current">
          Currently processing: <strong>{progress.current_system}</strong>
        </div>
      )}
    </div>
  );
}

export function PullProgress({ job, onCancel }: PullProgressProps) {
  const overallProgress = calculatePullProgress(job.progress);
  const isActive = isPullJobActive(job.status);
  const hasErrors = job.progress.errors && job.progress.errors.length > 0;

  return (
    <div className={`pull-progress ${job.status}`}>
      <div className="pull-progress-header">
        <span className="status-icon">{getStatusIcon(job.status)}</span>
        <span className="status-label">{getStatusLabel(job.status)}</span>
        {isActive && (
          <span className="overall-progress">{overallProgress}%</span>
        )}
      </div>

      {(isActive || job.status === 'completed') && (
        <ProgressStats progress={job.progress} />
      )}

      {job.error && (
        <div className="pull-progress-error alert alert-error">
          <strong>Error:</strong> {job.error}
        </div>
      )}

      {hasErrors && (
        <div className="pull-progress-warnings">
          <details>
            <summary>
              {job.progress.errors!.length} warning(s) during pull
            </summary>
            <ul>
              {job.progress.errors!.map((error, index) => (
                <li key={index}>{error}</li>
              ))}
            </ul>
          </details>
        </div>
      )}

      {isActive && onCancel && (
        <div className="pull-progress-actions">
          <button
            type="button"
            className="btn btn-secondary"
            onClick={onCancel}
          >
            Cancel Pull
          </button>
        </div>
      )}

      {job.completed_at && (
        <div className="pull-progress-timing">
          Completed at: {new Date(job.completed_at).toLocaleString()}
        </div>
      )}
    </div>
  );
}

export default PullProgress;
