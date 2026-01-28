import { useAuditEvent } from '../hooks/useAudit';
import { getEventTypeInfo, getStatusInfo } from '../types';

interface AuditDetailProps {
  eventId: string;
  onClose: () => void;
}

export function AuditDetail({ eventId, onClose }: AuditDetailProps) {
  const { data, isLoading, error } = useAuditEvent(eventId);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const formatDetails = (details: Record<string, unknown> | null) => {
    if (!details) return 'No additional details';
    return JSON.stringify(details, null, 2);
  };

  return (
    <div className="audit-detail-overlay" onClick={onClose}>
      <div className="audit-detail-modal" onClick={(e) => e.stopPropagation()}>
        <div className="audit-detail-header">
          <h2>Audit Event Details</h2>
          <button
            type="button"
            className="audit-detail-close"
            onClick={onClose}
            aria-label="Close"
          >
            ×
          </button>
        </div>

        {isLoading && (
          <div className="audit-detail-loading">Loading event details...</div>
        )}

        {error && (
          <div className="audit-detail-error">
            Failed to load event details. Please try again.
          </div>
        )}

        {data && (
          <div className="audit-detail-content">
            <div className="audit-detail-section">
              <h3>Event Information</h3>
              <dl className="audit-detail-grid">
                <dt>Event ID</dt>
                <dd className="audit-detail-id">{data.id}</dd>

                <dt>Event Type</dt>
                <dd>
                  <span
                    className="audit-badge"
                    style={{
                      color: getEventTypeInfo(data.event_type).color,
                      backgroundColor: getEventTypeInfo(data.event_type).bgColor,
                    }}
                  >
                    {getEventTypeInfo(data.event_type).label}
                  </span>
                </dd>

                <dt>Status</dt>
                <dd>
                  <span
                    className="audit-badge"
                    style={{
                      color: getStatusInfo(data.status).color,
                      backgroundColor: getStatusInfo(data.status).bgColor,
                    }}
                  >
                    {getStatusInfo(data.status).label}
                  </span>
                </dd>

                <dt>Timestamp</dt>
                <dd>{formatDate(data.created_at)}</dd>
              </dl>
            </div>

            <div className="audit-detail-section">
              <h3>Entity Information</h3>
              <dl className="audit-detail-grid">
                <dt>Entity Type</dt>
                <dd>{data.entity_type}</dd>

                <dt>Entity ID</dt>
                <dd className="audit-detail-id">{data.entity_id}</dd>

                <dt>Action</dt>
                <dd>{data.action}</dd>
              </dl>
            </div>

            <div className="audit-detail-section">
              <h3>User Information</h3>
              <dl className="audit-detail-grid">
                <dt>User</dt>
                <dd>{data.user_email || 'System'}</dd>

                <dt>IP Address</dt>
                <dd>{data.ip_address || 'N/A'}</dd>
              </dl>
            </div>

            {data.details && (
              <div className="audit-detail-section">
                <h3>Additional Details</h3>
                <pre className="audit-detail-json">
                  {formatDetails(data.details)}
                </pre>
              </div>
            )}
          </div>
        )}

        <div className="audit-detail-footer">
          <button type="button" className="btn btn-secondary" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </div>
  );
}

export default AuditDetail;
