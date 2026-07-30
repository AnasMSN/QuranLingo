import { create, isAxiosError, type AxiosError, type InternalAxiosRequestConfig } from 'axios';

import { API_URL } from './config';
import { clearTokens, loadTokens, saveTokens, type StoredTokens } from './tokenStorage';
import type {
  AuthResponse,
  CourseTree,
  LeaderboardResponse,
  LessonDetail,
  AnswerInput,
  SubmitResult,
  UserResponse,
  TokenPair,
} from '../types/api';

// In-memory mirror of the tokens in secure storage, so every request doesn't
// need an async keychain read. Hydrated once at app start via hydrateAuth().
let accessToken: string | null = null;
let refreshToken: string | null = null;

// Fires when the refresh token itself is invalid/expired — the auth store
// registers this to force a clean logout. Kept as a callback (rather than
// importing the store here) to avoid a client <-> store circular dependency.
let onAuthFailure: (() => void) | null = null;
export function setOnAuthFailure(cb: () => void) {
  onAuthFailure = cb;
}

export async function hydrateAuth(): Promise<UserResponse | null> {
  const tokens = await loadTokens();
  if (!tokens) return null;
  accessToken = tokens.accessToken;
  refreshToken = tokens.refreshToken;
  try {
    return await api.me();
  } catch {
    await applyLogout();
    return null;
  }
}

async function applyTokens(tokens: TokenPair): Promise<void> {
  accessToken = tokens.access_token;
  refreshToken = tokens.refresh_token;
  await saveTokens({ accessToken: tokens.access_token, refreshToken: tokens.refresh_token } satisfies StoredTokens);
}

async function applyLogout(): Promise<void> {
  accessToken = null;
  refreshToken = null;
  await clearTokens();
}

const http = create({ baseURL: API_URL, timeout: 15_000 });
// Separate, interceptor-free instance for the refresh call itself — it must
// never be caught by the 401 retry logic below (that would recurse forever).
const rawHttp = create({ baseURL: API_URL, timeout: 15_000 });

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  if (accessToken && !config.headers.Authorization) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

// Single-flight refresh: concurrent 401s while a refresh is in-flight all
// await the same promise instead of each firing their own /auth/refresh call.
let refreshPromise: Promise<string> | null = null;

async function refreshAccessToken(): Promise<string> {
  if (!refreshToken) throw new Error('no refresh token');
  if (!refreshPromise) {
    refreshPromise = rawHttp
      .post<{ tokens: TokenPair }>('/auth/refresh', { refresh_token: refreshToken })
      .then(async (res) => {
        await applyTokens(res.data.tokens);
        return res.data.tokens.access_token;
      })
      .finally(() => {
        refreshPromise = null;
      });
  }
  return refreshPromise;
}

http.interceptors.response.use(
  (res) => res,
  async (error: AxiosError) => {
    const original = error.config as (InternalAxiosRequestConfig & { _retried?: boolean }) | undefined;
    const isAuthEndpoint = original?.url?.includes('/auth/');

    if (error.response?.status === 401 && original && !original._retried && !isAuthEndpoint) {
      original._retried = true;
      try {
        const newAccessToken = await refreshAccessToken();
        original.headers.Authorization = `Bearer ${newAccessToken}`;
        return http.request(original);
      } catch (refreshError) {
        await applyLogout();
        onAuthFailure?.();
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  },
);

export function apiErrorMessage(error: unknown, fallback = 'Something went wrong'): string {
  if (isAxiosError(error)) {
    const body = error.response?.data as { error?: string } | undefined;
    if (body?.error) return body.error;
    if (error.message) return error.message;
  }
  return fallback;
}

export const api = {
  async register(email: string, password: string, displayName: string): Promise<AuthResponse> {
    const res = await http.post<AuthResponse>('/auth/register', {
      email,
      password,
      display_name: displayName,
    });
    await applyTokens(res.data.tokens);
    return res.data;
  },

  async login(email: string, password: string): Promise<AuthResponse> {
    const res = await http.post<AuthResponse>('/auth/login', { email, password });
    await applyTokens(res.data.tokens);
    return res.data;
  },

  async logout(): Promise<void> {
    await applyLogout();
  },

  async me(): Promise<UserResponse> {
    const res = await http.get<UserResponse>('/me');
    return res.data;
  },

  async getCourse(code: string): Promise<CourseTree> {
    const res = await http.get<CourseTree>(`/courses/${encodeURIComponent(code)}`);
    return res.data;
  },

  async getLesson(id: string): Promise<LessonDetail> {
    const res = await http.get<LessonDetail>(`/lessons/${encodeURIComponent(id)}`);
    return res.data;
  },

  async submitLesson(id: string, idempotencyKey: string, answers: AnswerInput[]): Promise<SubmitResult> {
    const res = await http.post<SubmitResult>(`/lessons/${encodeURIComponent(id)}/submit`, {
      idempotency_key: idempotencyKey,
      answers,
    });
    return res.data;
  },

  async leaderboardWeekly(): Promise<LeaderboardResponse> {
    const res = await http.get<LeaderboardResponse>('/leaderboard/weekly');
    return res.data;
  },
};
