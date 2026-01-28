// Policy Statement types for ServiceNow controls

export interface PolicyStatement {
  id: string;
  number: string;
  name: string;
  short_description: string;
  description?: string;
  state: string;
  category?: string;
  control_family?: string;
  created_at: string;
  updated_at: string;
}

export interface Pagination {
  page: number;
  page_size: number;
  total_count: number;
  total_pages: number;
}

export interface PolicyStatementsResponse {
  items: PolicyStatement[];
  pagination: Pagination;
}

export interface PolicyStatementsParams {
  page?: number;
  page_size?: number;
  search?: string;
  sort_by?: string;
  sort_dir?: 'asc' | 'desc';
}
