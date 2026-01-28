// Connection feature TypeScript types matching backend schemas

export type AuthMethod = 'basic' | 'oauth';

export type ConnectionStatus = 'success' | 'failure' | 'pending' | 'unknown';

// Request types
export interface ConfigRequest {
  instance_url: string;
  auth_method: AuthMethod;
  username?: string;
  password?: string;
  oauth_client_id?: string;
  oauth_client_secret?: string;
  oauth_token_url?: string;
}

// Response types
export interface StatusResponse {
  is_configured: boolean;
  instance_url?: string;
  auth_method?: AuthMethod;
  last_test_at?: string;
  last_test_status: ConnectionStatus;
  instance_version?: string;
}

export interface ConfigResponse {
  id: string;
  instance_url: string;
  auth_method: AuthMethod;
  status: ConnectionStatus;
  message: string;
}

export interface TestResponse {
  success: boolean;
  message?: string;
  instance_version?: string;
  build_tag?: string;
  response_time_ms?: number;
}

export interface ErrorResponse {
  error: string;
  message: string;
  details?: Record<string, string>;
}

export interface ValidationErrorResponse {
  error: string;
  message: string;
  fields?: ValidationError[];
}

export interface ValidationError {
  field: string;
  message: string;
}

// Form types for React Hook Form
export interface ConnectionFormData {
  instanceUrl: string;
  authMethod: AuthMethod;
  username: string;
  password: string;
  oauthClientId: string;
  oauthClientSecret: string;
  oauthTokenUrl: string;
}
