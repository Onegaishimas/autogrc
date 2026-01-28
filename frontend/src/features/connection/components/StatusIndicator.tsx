import { useConnectionStatus } from '../hooks/useConnection';
import type { ConnectionStatus } from '../types';

interface StatusBadgeProps {
  status: ConnectionStatus;
  isLoading?: boolean;
}

function StatusBadge({ status, isLoading }: StatusBadgeProps) {
  if (isLoading) {
    return (
      <span className="status-badge status-loading">
        <span className="status-dot loading" />
        Loading...
      </span>
    );
  }

  const statusConfig: Record<ConnectionStatus, { label: string; className: string }> = {
    success: { label: 'Connected', className: 'status-success' },
    failure: { label: 'Failed', className: 'status-failure' },
    pending: { label: 'Pending', className: 'status-pending' },
    unknown: { label: 'Not Configured', className: 'status-unknown' },
  };

  const config = statusConfig[status] || statusConfig.unknown;

  return (
    <span className={`status-badge ${config.className}`}>
      <span className="status-dot" />
      {config.label}
    </span>
  );
}

export function StatusIndicator() {
  const { data: status, isLoading, isError } = useConnectionStatus();

  if (isError) {
    return <StatusBadge status="failure" />;
  }

  if (isLoading || !status) {
    return <StatusBadge status="unknown" isLoading={isLoading} />;
  }

  if (!status.is_configured) {
    return <StatusBadge status="unknown" />;
  }

  return <StatusBadge status={status.last_test_status} />;
}

export default StatusIndicator;
