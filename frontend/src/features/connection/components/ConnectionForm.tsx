import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useSaveConnection } from '../hooks/useConnection';
import type { AuthMethod, ConfigRequest } from '../types';
import { getErrorMessage } from '../../../lib/api';

// Zod schema for form validation
const connectionSchema = z.object({
  instanceUrl: z
    .string()
    .min(1, 'Instance URL is required')
    .url('Must be a valid URL')
    .refine((url) => url.startsWith('https://'), 'URL must use HTTPS'),
  authMethod: z.enum(['basic', 'oauth'] as const),
  username: z.string().optional(),
  password: z.string().optional(),
  oauthClientId: z.string().optional(),
  oauthClientSecret: z.string().optional(),
  oauthTokenUrl: z.string().optional(),
}).superRefine((data, ctx) => {
  if (data.authMethod === 'basic') {
    if (!data.username || data.username.trim() === '') {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Username is required for basic authentication',
        path: ['username'],
      });
    }
    if (!data.password || data.password.trim() === '') {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Password is required for basic authentication',
        path: ['password'],
      });
    }
  } else if (data.authMethod === 'oauth') {
    if (!data.oauthClientId || data.oauthClientId.trim() === '') {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Client ID is required for OAuth',
        path: ['oauthClientId'],
      });
    }
    if (!data.oauthClientSecret || data.oauthClientSecret.trim() === '') {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Client Secret is required for OAuth',
        path: ['oauthClientSecret'],
      });
    }
    if (!data.oauthTokenUrl || data.oauthTokenUrl.trim() === '') {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Token URL is required for OAuth',
        path: ['oauthTokenUrl'],
      });
    }
  }
});

type ConnectionFormData = z.infer<typeof connectionSchema>;

interface ConnectionFormProps {
  initialData?: Partial<ConnectionFormData>;
  onSuccess?: () => void;
}

export function ConnectionForm({ initialData, onSuccess }: ConnectionFormProps) {
  const saveConnection = useSaveConnection();

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<ConnectionFormData>({
    resolver: zodResolver(connectionSchema),
    defaultValues: {
      instanceUrl: initialData?.instanceUrl || '',
      authMethod: initialData?.authMethod || 'basic',
      username: initialData?.username || '',
      password: '',
      oauthClientId: initialData?.oauthClientId || '',
      oauthClientSecret: '',
      oauthTokenUrl: initialData?.oauthTokenUrl || '',
    },
  });

  const authMethod = watch('authMethod');

  const onSubmit = async (data: ConnectionFormData) => {
    const request: ConfigRequest = {
      instance_url: data.instanceUrl,
      auth_method: data.authMethod as AuthMethod,
    };

    if (data.authMethod === 'basic') {
      request.username = data.username;
      request.password = data.password;
    } else {
      request.oauth_client_id = data.oauthClientId;
      request.oauth_client_secret = data.oauthClientSecret;
      request.oauth_token_url = data.oauthTokenUrl;
    }

    try {
      await saveConnection.mutateAsync(request);
      onSuccess?.();
    } catch (error) {
      // Error is handled by the mutation
      console.error('Failed to save connection:', getErrorMessage(error));
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="connection-form">
      <div className="form-group">
        <label htmlFor="instanceUrl">ServiceNow Instance URL</label>
        <input
          id="instanceUrl"
          type="url"
          placeholder="https://your-instance.service-now.com"
          {...register('instanceUrl')}
          className={errors.instanceUrl ? 'error' : ''}
        />
        {errors.instanceUrl && (
          <span className="error-message">{errors.instanceUrl.message}</span>
        )}
      </div>

      <div className="form-group">
        <label>Authentication Method</label>
        <div className="radio-group">
          <label className="radio-label">
            <input
              type="radio"
              value="basic"
              {...register('authMethod')}
            />
            Basic Authentication
          </label>
          <label className="radio-label">
            <input
              type="radio"
              value="oauth"
              {...register('authMethod')}
            />
            OAuth 2.0
          </label>
        </div>
      </div>

      {authMethod === 'basic' && (
        <>
          <div className="form-group">
            <label htmlFor="username">Username</label>
            <input
              id="username"
              type="text"
              placeholder="admin"
              {...register('username')}
              className={errors.username ? 'error' : ''}
            />
            {errors.username && (
              <span className="error-message">{errors.username.message}</span>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              type="password"
              placeholder="Enter password"
              {...register('password')}
              className={errors.password ? 'error' : ''}
            />
            {errors.password && (
              <span className="error-message">{errors.password.message}</span>
            )}
          </div>
        </>
      )}

      {authMethod === 'oauth' && (
        <>
          <div className="form-group">
            <label htmlFor="oauthClientId">Client ID</label>
            <input
              id="oauthClientId"
              type="text"
              placeholder="OAuth Client ID"
              {...register('oauthClientId')}
              className={errors.oauthClientId ? 'error' : ''}
            />
            {errors.oauthClientId && (
              <span className="error-message">{errors.oauthClientId.message}</span>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="oauthClientSecret">Client Secret</label>
            <input
              id="oauthClientSecret"
              type="password"
              placeholder="OAuth Client Secret"
              {...register('oauthClientSecret')}
              className={errors.oauthClientSecret ? 'error' : ''}
            />
            {errors.oauthClientSecret && (
              <span className="error-message">{errors.oauthClientSecret.message}</span>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="oauthTokenUrl">Token URL</label>
            <input
              id="oauthTokenUrl"
              type="url"
              placeholder="https://your-instance.service-now.com/oauth_token.do"
              {...register('oauthTokenUrl')}
              className={errors.oauthTokenUrl ? 'error' : ''}
            />
            {errors.oauthTokenUrl && (
              <span className="error-message">{errors.oauthTokenUrl.message}</span>
            )}
          </div>
        </>
      )}

      {saveConnection.isError && (
        <div className="form-error">
          {getErrorMessage(saveConnection.error)}
        </div>
      )}

      <div className="form-actions">
        <button
          type="submit"
          disabled={isSubmitting || saveConnection.isPending}
          className="btn btn-primary"
        >
          {saveConnection.isPending ? 'Saving...' : 'Save Configuration'}
        </button>
      </div>
    </form>
  );
}

export default ConnectionForm;
