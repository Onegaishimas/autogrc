import type { PushJob } from '../types';
import { getPushStatusInfo } from '../types';

interface PushProgressProps {
  job: PushJob;
  onCancel?: () => void;
}

export function PushProgress({ job, onCancel }: PushProgressProps) {
  const statusInfo = getPushStatusInfo(job.status);
  const progressPercent =
    job.total_count > 0
      ? Math.round((job.completed / job.total_count) * 100)
      : 0;

  return (
    <div className="push-progress">
      <div className="push-progress-header">
        <span
          className="push-progress-badge"
          style={{
            color: statusInfo.color,
            backgroundColor: statusInfo.bgColor,
          }}
        >
          {statusInfo.label}
        </span>
        {job.status === 'running' && onCancel && (
          <button
            type="button"
            className="btn btn-secondary btn-sm"
            onClick={onCancel}
          >
            Cancel
          </button>
        )}
      </div>

      <div className="push-progress-bar-container">
        <div className="push-progress-bar">
          <div
            className="push-progress-fill"
            style={{ width: `${progressPercent}%` }}
          />
        </div>
        <span className="push-progress-percent">{progressPercent}%</span>
      </div>

      <div className="push-progress-stats">
        <div className="push-progress-stat">
          <span className="stat-value">{job.completed}</span>
          <span className="stat-label">/ {job.total_count} completed</span>
        </div>
        <div className="push-progress-stat success">
          <span className="stat-value">{job.succeeded}</span>
          <span className="stat-label">succeeded</span>
        </div>
        {job.failed > 0 && (
          <div className="push-progress-stat error">
            <span className="stat-value">{job.failed}</span>
            <span className="stat-label">failed</span>
          </div>
        )}
      </div>

      {job.status === 'running' && (
        <p className="push-progress-message">
          Pushing statements to ServiceNow...
        </p>
      )}
    </div>
  );
}

export default PushProgress;
