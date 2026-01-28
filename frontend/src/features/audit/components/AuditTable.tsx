import type { AuditEvent } from '../types';
import { getEventTypeInfo, getStatusInfo } from '../types';

interface AuditTableProps {
  events: AuditEvent[];
  isLoading: boolean;
  onSelectEvent: (id: string) => void;
}

export function AuditTable({ events, isLoading, onSelectEvent }: AuditTableProps) {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const truncate = (str: string, maxLength: number) => {
    if (str.length <= maxLength) return str;
    return str.slice(0, maxLength) + '...';
  };

  if (isLoading) {
    return <div className="audit-table-loading">Loading audit events...</div>;
  }

  if (events.length === 0) {
    return (
      <div className="audit-table-empty">
        <p>No audit events found.</p>
        <p className="text-muted">
          Events will appear here as you use the application.
        </p>
      </div>
    );
  }

  return (
    <div className="audit-table-container">
      <table className="audit-table">
        <thead>
          <tr>
            <th>Time</th>
            <th>Event</th>
            <th>Entity</th>
            <th>Action</th>
            <th>Status</th>
            <th>User</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {events.map((event) => {
            const eventInfo = getEventTypeInfo(event.event_type);
            const statusInfo = getStatusInfo(event.status);

            return (
              <tr key={event.id}>
                <td className="audit-table-time">{formatDate(event.created_at)}</td>
                <td>
                  <span
                    className="audit-badge"
                    style={{
                      color: eventInfo.color,
                      backgroundColor: eventInfo.bgColor,
                    }}
                  >
                    {eventInfo.label}
                  </span>
                </td>
                <td className="audit-table-entity">
                  <span className="audit-entity-type">{event.entity_type}:</span>
                  <span className="audit-entity-id">
                    {truncate(event.entity_id, 20)}
                  </span>
                </td>
                <td>{truncate(event.action, 30)}</td>
                <td>
                  <span
                    className="audit-badge"
                    style={{
                      color: statusInfo.color,
                      backgroundColor: statusInfo.bgColor,
                    }}
                  >
                    {statusInfo.label}
                  </span>
                </td>
                <td className="audit-table-user">
                  {event.user_email || 'System'}
                </td>
                <td>
                  <button
                    type="button"
                    className="btn btn-secondary btn-sm"
                    onClick={() => onSelectEvent(event.id)}
                  >
                    View
                  </button>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

export default AuditTable;
