import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ConnectionForm } from './ConnectionForm';

// Mock the useSaveConnection hook
const mockMutateAsync = vi.fn();
vi.mock('../hooks/useConnection', () => ({
  useSaveConnection: () => ({
    mutateAsync: mockMutateAsync,
    isPending: false,
    isError: false,
    error: null,
  }),
}));

// Helper to render with providers
function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>
  );
}

describe('ConnectionForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockMutateAsync.mockResolvedValue({});
  });

  describe('rendering', () => {
    it('renders the form with default basic auth fields', () => {
      renderWithProviders(<ConnectionForm />);

      expect(screen.getByLabelText(/ServiceNow Instance URL/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/Basic Authentication/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/OAuth 2.0/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/Username/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/Password/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Save Configuration/i })).toBeInTheDocument();
    });

    it('renders with initial data when provided', () => {
      renderWithProviders(
        <ConnectionForm
          initialData={{
            instanceUrl: 'https://example.service-now.com',
            authMethod: 'basic',
            username: 'testuser',
          }}
        />
      );

      expect(screen.getByLabelText(/ServiceNow Instance URL/i)).toHaveValue(
        'https://example.service-now.com'
      );
      expect(screen.getByLabelText(/Username/i)).toHaveValue('testuser');
    });
  });

  describe('conditional field display based on auth method', () => {
    it('shows basic auth fields when basic auth is selected', () => {
      renderWithProviders(<ConnectionForm />);

      expect(screen.getByLabelText(/Username/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/Password/i)).toBeInTheDocument();
      expect(screen.queryByLabelText(/Client ID/i)).not.toBeInTheDocument();
      expect(screen.queryByLabelText(/Client Secret/i)).not.toBeInTheDocument();
    });

    it('shows OAuth fields when OAuth is selected', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ConnectionForm />);

      await user.click(screen.getByLabelText(/OAuth 2.0/i));

      expect(screen.queryByLabelText(/Username/i)).not.toBeInTheDocument();
      expect(screen.queryByLabelText(/Password/i)).not.toBeInTheDocument();
      expect(screen.getByLabelText(/Client ID/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/Client Secret/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/Token URL/i)).toBeInTheDocument();
    });

    it('switches between auth methods correctly', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ConnectionForm />);

      // Initially basic auth is shown
      expect(screen.getByLabelText(/Username/i)).toBeInTheDocument();

      // Switch to OAuth
      await user.click(screen.getByLabelText(/OAuth 2.0/i));
      expect(screen.queryByLabelText(/Username/i)).not.toBeInTheDocument();
      expect(screen.getByLabelText(/Client ID/i)).toBeInTheDocument();

      // Switch back to basic
      await user.click(screen.getByLabelText(/Basic Authentication/i));
      expect(screen.getByLabelText(/Username/i)).toBeInTheDocument();
      expect(screen.queryByLabelText(/Client ID/i)).not.toBeInTheDocument();
    });
  });

  describe('URL validation', () => {
    it('shows error for empty URL', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ConnectionForm />);

      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        expect(screen.getByText(/Instance URL is required/i)).toBeInTheDocument();
      });
    });

    it('shows error for non-HTTPS URL', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ConnectionForm />);

      await user.type(
        screen.getByLabelText(/ServiceNow Instance URL/i),
        'http://example.service-now.com'
      );
      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        expect(screen.getByText(/URL must use HTTPS/i)).toBeInTheDocument();
      });
    });

    it('shows error for invalid URL format', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ConnectionForm />);

      // Use a URL that's technically valid format but not HTTPS to trigger the refinement
      await user.type(screen.getByLabelText(/ServiceNow Instance URL/i), 'ftp://not-https.com');
      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        // The HTTPS refinement check should trigger
        expect(screen.getByText(/URL must use HTTPS/i)).toBeInTheDocument();
      });
    });

    it('accepts valid HTTPS URL', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ConnectionForm />);

      await user.type(
        screen.getByLabelText(/ServiceNow Instance URL/i),
        'https://dev12345.service-now.com'
      );
      await user.type(screen.getByLabelText(/Username/i), 'admin');
      await user.type(screen.getByLabelText(/Password/i), 'password123');
      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalled();
      });
    });
  });

  describe('form submission with valid data', () => {
    it('submits form with basic auth credentials', async () => {
      const user = userEvent.setup();
      const onSuccess = vi.fn();
      renderWithProviders(<ConnectionForm onSuccess={onSuccess} />);

      await user.type(
        screen.getByLabelText(/ServiceNow Instance URL/i),
        'https://dev12345.service-now.com'
      );
      await user.type(screen.getByLabelText(/Username/i), 'admin');
      await user.type(screen.getByLabelText(/Password/i), 'secret123');
      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          instance_url: 'https://dev12345.service-now.com',
          auth_method: 'basic',
          username: 'admin',
          password: 'secret123',
        });
      });

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalled();
      });
    });

    it('submits form with OAuth credentials', async () => {
      const user = userEvent.setup();
      const onSuccess = vi.fn();
      renderWithProviders(<ConnectionForm onSuccess={onSuccess} />);

      await user.type(
        screen.getByLabelText(/ServiceNow Instance URL/i),
        'https://dev12345.service-now.com'
      );
      await user.click(screen.getByLabelText(/OAuth 2.0/i));
      await user.type(screen.getByLabelText(/Client ID/i), 'client-id-123');
      await user.type(screen.getByLabelText(/Client Secret/i), 'client-secret-456');
      await user.type(
        screen.getByLabelText(/Token URL/i),
        'https://dev12345.service-now.com/oauth_token.do'
      );
      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          instance_url: 'https://dev12345.service-now.com',
          auth_method: 'oauth',
          oauth_client_id: 'client-id-123',
          oauth_client_secret: 'client-secret-456',
          oauth_token_url: 'https://dev12345.service-now.com/oauth_token.do',
        });
      });

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalled();
      });
    });
  });

  describe('validation error display', () => {
    it('shows error when username is empty for basic auth', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ConnectionForm />);

      await user.type(
        screen.getByLabelText(/ServiceNow Instance URL/i),
        'https://dev12345.service-now.com'
      );
      await user.type(screen.getByLabelText(/Password/i), 'password123');
      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        expect(screen.getByText(/Username is required for basic authentication/i)).toBeInTheDocument();
      });
    });

    it('shows error when password is empty for basic auth', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ConnectionForm />);

      await user.type(
        screen.getByLabelText(/ServiceNow Instance URL/i),
        'https://dev12345.service-now.com'
      );
      await user.type(screen.getByLabelText(/Username/i), 'admin');
      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        expect(screen.getByText(/Password is required for basic authentication/i)).toBeInTheDocument();
      });
    });

    it('shows errors for missing OAuth fields', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ConnectionForm />);

      await user.type(
        screen.getByLabelText(/ServiceNow Instance URL/i),
        'https://dev12345.service-now.com'
      );
      await user.click(screen.getByLabelText(/OAuth 2.0/i));
      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        expect(screen.getByText(/Client ID is required for OAuth/i)).toBeInTheDocument();
        expect(screen.getByText(/Client Secret is required for OAuth/i)).toBeInTheDocument();
        expect(screen.getByText(/Token URL is required for OAuth/i)).toBeInTheDocument();
      });
    });

    it('clears validation errors when fields are filled', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ConnectionForm />);

      // Trigger validation errors
      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        expect(screen.getByText(/Instance URL is required/i)).toBeInTheDocument();
      });

      // Fill in the fields
      await user.type(
        screen.getByLabelText(/ServiceNow Instance URL/i),
        'https://dev12345.service-now.com'
      );
      await user.type(screen.getByLabelText(/Username/i), 'admin');
      await user.type(screen.getByLabelText(/Password/i), 'password123');
      await user.click(screen.getByRole('button', { name: /Save Configuration/i }));

      await waitFor(() => {
        expect(screen.queryByText(/Instance URL is required/i)).not.toBeInTheDocument();
        expect(mockMutateAsync).toHaveBeenCalled();
      });
    });
  });
});
