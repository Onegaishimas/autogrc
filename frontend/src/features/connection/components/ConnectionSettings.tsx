import { useState } from 'react';
import { useConnectionStatus, useDeleteConnection } from '../hooks/useConnection';
import { ConnectionForm } from './ConnectionForm';
import { ConnectionTest } from './ConnectionTest';
import { StatusIndicator } from './StatusIndicator';
import { getErrorMessage } from '../../../lib/api';

export function ConnectionSettings() {
  const { data: status, isLoading } = useConnectionStatus();
  const deleteConnection = useDeleteConnection();
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const handleDelete = async () => {
    try {
      await deleteConnection.mutateAsync();
      setShowDeleteConfirm(false);
    } catch (error) {
      console.error('Failed to delete connection:', getErrorMessage(error));
    }
  };

  return (
    <div className="connection-settings">
      <header className="settings-header">
        <div className="header-content">
          <h1>ServiceNow Connection</h1>
          <StatusIndicator />
        </div>
        <p className="settings-description">
          Configure your ServiceNow GRC instance connection to enable bidirectional
          synchronization of control implementation statements.
        </p>
      </header>

      <main className="settings-content">
        <section className="settings-section">
          <h2>Connection Configuration</h2>
          {isLoading ? (
            <div className="loading">Loading configuration...</div>
          ) : (
            <ConnectionForm
              initialData={
                status?.is_configured
                  ? {
                      instanceUrl: status.instance_url,
                      authMethod: status.auth_method,
                    }
                  : undefined
              }
              onSuccess={() => {
                // Configuration saved successfully
              }}
            />
          )}
        </section>

        <section className="settings-section">
          <ConnectionTest />
        </section>

        {status?.is_configured && (
          <section className="settings-section danger-zone">
            <h2>Danger Zone</h2>
            <div className="danger-content">
              <div className="danger-info">
                <h3>Delete Connection</h3>
                <p>
                  Remove the ServiceNow connection configuration. This will delete
                  stored credentials and disable synchronization.
                </p>
              </div>
              {!showDeleteConfirm ? (
                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  className="btn btn-danger"
                >
                  Delete Connection
                </button>
              ) : (
                <div className="confirm-delete">
                  <p>Are you sure you want to delete the connection?</p>
                  <div className="confirm-actions">
                    <button
                      onClick={handleDelete}
                      disabled={deleteConnection.isPending}
                      className="btn btn-danger"
                    >
                      {deleteConnection.isPending ? 'Deleting...' : 'Yes, Delete'}
                    </button>
                    <button
                      onClick={() => setShowDeleteConfirm(false)}
                      className="btn btn-secondary"
                    >
                      Cancel
                    </button>
                  </div>
                  {deleteConnection.isError && (
                    <p className="error-message">
                      {getErrorMessage(deleteConnection.error)}
                    </p>
                  )}
                </div>
              )}
            </div>
          </section>
        )}
      </main>
    </div>
  );
}

export default ConnectionSettings;
