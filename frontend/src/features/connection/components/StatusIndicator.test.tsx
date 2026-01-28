import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { StatusIndicator } from './StatusIndicator';

// Mock the useConnectionStatus hook
const mockUseConnectionStatus = vi.fn();
vi.mock('../hooks/useConnection', () => ({
  useConnectionStatus: () => mockUseConnectionStatus(),
}));

// Helper to render with providers
function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>
  );
}

describe('StatusIndicator', () => {
  describe('loading state', () => {
    it('shows loading indicator when data is loading', () => {
      mockUseConnectionStatus.mockReturnValue({
        data: undefined,
        isLoading: true,
        isError: false,
      });

      renderWithProviders(<StatusIndicator />);

      expect(screen.getByText('Loading...')).toBeInTheDocument();
      expect(document.querySelector('.status-loading')).toBeInTheDocument();
    });

    it('shows loading indicator with animated dot', () => {
      mockUseConnectionStatus.mockReturnValue({
        data: undefined,
        isLoading: true,
        isError: false,
      });

      renderWithProviders(<StatusIndicator />);

      const loadingDot = document.querySelector('.status-dot.loading');
      expect(loadingDot).toBeInTheDocument();
    });
  });

  describe('status badge states', () => {
    it('shows "Connected" with success class when status is success', () => {
      mockUseConnectionStatus.mockReturnValue({
        data: {
          is_configured: true,
          last_test_status: 'success',
        },
        isLoading: false,
        isError: false,
      });

      renderWithProviders(<StatusIndicator />);

      expect(screen.getByText('Connected')).toBeInTheDocument();
      expect(document.querySelector('.status-success')).toBeInTheDocument();
    });

    it('shows "Failed" with failure class when status is failure', () => {
      mockUseConnectionStatus.mockReturnValue({
        data: {
          is_configured: true,
          last_test_status: 'failure',
        },
        isLoading: false,
        isError: false,
      });

      renderWithProviders(<StatusIndicator />);

      expect(screen.getByText('Failed')).toBeInTheDocument();
      expect(document.querySelector('.status-failure')).toBeInTheDocument();
    });

    it('shows "Pending" with pending class when status is pending', () => {
      mockUseConnectionStatus.mockReturnValue({
        data: {
          is_configured: true,
          last_test_status: 'pending',
        },
        isLoading: false,
        isError: false,
      });

      renderWithProviders(<StatusIndicator />);

      expect(screen.getByText('Pending')).toBeInTheDocument();
      expect(document.querySelector('.status-pending')).toBeInTheDocument();
    });

    it('shows "Not Configured" with unknown class when status is unknown', () => {
      mockUseConnectionStatus.mockReturnValue({
        data: {
          is_configured: true,
          last_test_status: 'unknown',
        },
        isLoading: false,
        isError: false,
      });

      renderWithProviders(<StatusIndicator />);

      expect(screen.getByText('Not Configured')).toBeInTheDocument();
      expect(document.querySelector('.status-unknown')).toBeInTheDocument();
    });
  });

  describe('error handling', () => {
    it('shows failure badge when there is an error', () => {
      mockUseConnectionStatus.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
      });

      renderWithProviders(<StatusIndicator />);

      expect(screen.getByText('Failed')).toBeInTheDocument();
      expect(document.querySelector('.status-failure')).toBeInTheDocument();
    });
  });

  describe('not configured state', () => {
    it('shows "Not Configured" when connection is not configured', () => {
      mockUseConnectionStatus.mockReturnValue({
        data: {
          is_configured: false,
          last_test_status: 'unknown',
        },
        isLoading: false,
        isError: false,
      });

      renderWithProviders(<StatusIndicator />);

      expect(screen.getByText('Not Configured')).toBeInTheDocument();
      expect(document.querySelector('.status-unknown')).toBeInTheDocument();
    });

    it('shows "Not Configured" when data is null', () => {
      mockUseConnectionStatus.mockReturnValue({
        data: null,
        isLoading: false,
        isError: false,
      });

      renderWithProviders(<StatusIndicator />);

      expect(screen.getByText('Not Configured')).toBeInTheDocument();
    });
  });

  describe('status dot', () => {
    it('renders status dot for all states', () => {
      mockUseConnectionStatus.mockReturnValue({
        data: {
          is_configured: true,
          last_test_status: 'success',
        },
        isLoading: false,
        isError: false,
      });

      renderWithProviders(<StatusIndicator />);

      expect(document.querySelector('.status-dot')).toBeInTheDocument();
    });
  });
});
