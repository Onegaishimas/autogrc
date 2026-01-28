// Types for system management

export interface DiscoveredSystem {
  sn_sys_id: string;
  name: string;
  description?: string;
  owner?: string;
  is_imported: boolean;
}

export interface DiscoverSystemsResponse {
  systems: DiscoveredSystem[];
  count: number;
}

export interface LocalSystem {
  id: string;
  sn_sys_id: string;
  name: string;
  description?: string;
  acronym?: string;
  owner?: string;
  status: string;
  control_count: number;
  statement_count: number;
  modified_count: number;
  last_pull_at?: string;
  last_push_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ListSystemsResponse {
  systems: LocalSystem[];
  total_count: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ListSystemsParams {
  page?: number;
  page_size?: number;
  search?: string;
  status?: string;
}

export interface ImportSystemsRequest {
  sn_sys_ids: string[];
}

export interface ImportSystemsResponse {
  imported: LocalSystem[];
  count: number;
}
