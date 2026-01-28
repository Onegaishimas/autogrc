import type { Pagination } from '../types';

interface ControlsPaginationProps {
  pagination: Pagination;
  onPageChange: (page: number) => void;
}

export function ControlsPagination({ pagination, onPageChange }: ControlsPaginationProps) {
  const { page, total_pages, total_count, page_size } = pagination;

  const startItem = (page - 1) * page_size + 1;
  const endItem = Math.min(page * page_size, total_count);

  const canGoPrev = page > 1;
  const canGoNext = page < total_pages;

  return (
    <div className="controls-pagination">
      <div className="pagination-info">
        Showing {startItem}-{endItem} of {total_count} items
      </div>
      <div className="pagination-controls">
        <button
          onClick={() => onPageChange(1)}
          disabled={!canGoPrev}
          className="btn btn-secondary pagination-btn"
          aria-label="First page"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="m11 17-5-5 5-5" />
            <path d="m18 17-5-5 5-5" />
          </svg>
        </button>
        <button
          onClick={() => onPageChange(page - 1)}
          disabled={!canGoPrev}
          className="btn btn-secondary pagination-btn"
          aria-label="Previous page"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="m15 18-6-6 6-6" />
          </svg>
        </button>
        <span className="pagination-page">
          Page {page} of {total_pages}
        </span>
        <button
          onClick={() => onPageChange(page + 1)}
          disabled={!canGoNext}
          className="btn btn-secondary pagination-btn"
          aria-label="Next page"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="m9 18 6-6-6-6" />
          </svg>
        </button>
        <button
          onClick={() => onPageChange(total_pages)}
          disabled={!canGoNext}
          className="btn btn-secondary pagination-btn"
          aria-label="Last page"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="m6 17 5-5-5-5" />
            <path d="m13 17 5-5-5-5" />
          </svg>
        </button>
      </div>
    </div>
  );
}
