import { useTestConnection } from '../hooks/useConnectionTest';
import { useConnectionStatus } from '../hooks/useConnection';
import { getErrorMessage } from '../../../lib/api';

export function ConnectionTest() {
  const { data: status } = useConnectionStatus();
  const testConnection = useTestConnection();

  const handleTest = () => {
    testConnection.mutate();
  };

  const isConfigured = status?.is_configured ?? false;

  return (
    <div className="connection-test">
      <div className="test-header">
        <h3>Test Connection</h3>
        <button
          onClick={handleTest}
          disabled={!isConfigured || testConnection.isPending}
          className="btn btn-secondary"
        >
          {testConnection.isPending ? 'Testing...' : 'Test Connection'}
        </button>
      </div>

      {!isConfigured && (
        <p className="test-info">
          Configure your ServiceNow connection above before testing.
        </p>
      )}

      {testConnection.isError && (
        <div className="test-result error">
          <strong>Test Failed</strong>
          <p>{getErrorMessage(testConnection.error)}</p>
        </div>
      )}

      {testConnection.isSuccess && testConnection.data && (
        <div className={`test-result ${testConnection.data.success ? 'success' : 'failure'}`}>
          <strong>
            {testConnection.data.success ? 'Connection Successful' : 'Connection Failed'}
          </strong>

          {testConnection.data.success ? (
            <div className="test-details">
              {testConnection.data.instance_version && (
                <p>
                  <span className="label">Instance Version:</span>
                  <span className="value">{testConnection.data.instance_version}</span>
                </p>
              )}
              {testConnection.data.build_tag && (
                <p>
                  <span className="label">Build Tag:</span>
                  <span className="value">{testConnection.data.build_tag}</span>
                </p>
              )}
              {testConnection.data.response_time_ms !== undefined && (
                <p>
                  <span className="label">Response Time:</span>
                  <span className="value">{testConnection.data.response_time_ms}ms</span>
                </p>
              )}
            </div>
          ) : (
            <p className="error-message">{testConnection.data.message}</p>
          )}
        </div>
      )}

      {status?.last_test_at && !testConnection.isPending && !testConnection.isSuccess && (
        <div className="last-test-info">
          <p>
            <span className="label">Last Tested:</span>
            <span className="value">
              {new Date(status.last_test_at).toLocaleString()}
            </span>
          </p>
          <p>
            <span className="label">Status:</span>
            <span className={`value status-${status.last_test_status}`}>
              {status.last_test_status}
            </span>
          </p>
          {status.instance_version && (
            <p>
              <span className="label">Version:</span>
              <span className="value">{status.instance_version}</span>
            </p>
          )}
        </div>
      )}
    </div>
  );
}

export default ConnectionTest;
